package dbops

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/tpillz-presents/service/store-api/store"
)

func TestPutStoreItem(t *testing.T) {
	var tests = []struct {
		category       string
		subcat         string
		itemID         string
		name           string
		price          float32
		unitsSold      int
		unitsAvailable map[string]int
	}{
		{category: "games", subcat: "game_sets", itemID: "001", name: "PawnWars Game Set", price: 29.95, unitsSold: 0, unitsAvailable: map[string]int{"OS": 10}},
		{category: "games", subcat: "game_sets", itemID: "002", name: "PawnWars Chess Pieces", price: 14.95, unitsSold: 0, unitsAvailable: map[string]int{"OS": 10}},
		{category: "artwork", subcat: "posters", itemID: "003", name: "PawnWars King Poster", price: 17.95, unitsSold: 0, unitsAvailable: map[string]int{"OS": 20}},
		{category: "artwork", subcat: "posters", itemID: "004", name: "PawnWars Queen Poster", price: 17.95, unitsSold: 0, unitsAvailable: map[string]int{"OS": 20}},
		{category: "clothing", subcat: "shirts", itemID: "005", name: "ACamoPrjct Logo T-Shirt", price: 22.95, unitsSold: 0, unitsAvailable: map[string]int{"S": 10, "M": 8, "L": 6, "XL": 4}},
	}

	// init db
	table := NewTable(StoreItemsTable, StoreItemPK, StoreItemSK)
	tables := []Table{table}
	dbInfo := InitDB(tables)

	for _, test := range tests {
		item := &store.StoreItem{
			ItemID:         test.itemID,
			Name:           test.name,
			Category:       test.category,
			Subcategory:    test.subcat,
			Price:          test.price,
			UnitsSold:      test.unitsSold,
			UnitsAvailable: test.unitsAvailable,
		}
		err := PutStoreItem(dbInfo, item)
		if err != nil {
			t.Errorf("FAIL: %v; want: nil", err)
		}
	}
}

func TestGetStoreItem(t *testing.T) {
	var tests = []struct {
		subcat   string
		itemID   string
		wantName string
	}{
		{subcat: "game_sets", itemID: "001", wantName: "PawnWars Game Set"},
		{subcat: "posters", itemID: "003", wantName: "PawnWars King Poster"},
		{subcat: "shirts", itemID: "005", wantName: "ACamoPrjct Logo T-Shirt"},
		{subcat: "pants", itemID: "006", wantName: ""},  // non existent partition
		{subcat: "shirts", itemID: "007", wantName: ""}, // non existent item
	}

	table := NewTable(StoreItemsTable, StoreItemPK, StoreItemSK)
	tables := []Table{table}
	dbInfo := InitDB(tables)

	for _, test := range tests {
		item, err := GetStoreItem(dbInfo, test.subcat, test.itemID)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
		if item.Name != test.wantName {
			t.Errorf("FAIL - DATA: %s; want: %v", item.Name, test.wantName)
		}
	}
}

func TestCheckInventory(t *testing.T) {
	var tests = []struct {
		item *store.CartItem
		want bool
	}{
		{item: &store.CartItem{Subcategory: "game_sets", ItemID: "001", Size: "OS", Quantity: 1}, want: true}, // in stock
		{item: &store.CartItem{Subcategory: "shirts", ItemID: "005", Size: "XL", Quantity: 5}, want: false},   // Insufficient stock
		{item: &store.CartItem{Subcategory: "shirts", ItemID: "009", Size: "M", Quantity: 1}, want: false},    // Non existent item
		{item: &store.CartItem{Subcategory: "pants", ItemID: "010", Size: "32", Quantity: 1}, want: false},    // Non existent partition
	}

	table := NewTable(StoreItemsTable, StoreItemPK, StoreItemSK)
	tables := []Table{table}
	dbInfo := InitDB(tables)
	bc := make(chan map[string]bool)
	ec := make(chan error)
	var wg sync.WaitGroup

	wantMap := make(map[string]bool)

	for _, test := range tests {
		wg.Add(1)
		wantMap[test.item.ItemID] = test.want
		go checkInventory(dbInfo, test.item, bc, ec, &wg)
	}

	br := 0
	er := 0
	for {
		select {
		case check := <-bc:
			br++
			t.Logf("bool received: %v", check)
			for k, v := range check {
				if wantMap[k] != v {
					t.Errorf("FAIL: %v; want: %v", v, wantMap[k])
				}
			}
		case err := <-ec:
			er++
			t.Logf("error received: %v", err)
			if err != nil {
				t.Errorf("FAIL: %v", err)
			}
		}
		if er == len(tests) && br == len(tests) {
			break
		}
		log.Println()
	}

	wg.Wait()
	close(bc)
	close(ec)
	t.Log("chans closed")

}

func TestUpdateInventoryCount(t *testing.T) {
	var tests = []struct {
		subcat  string
		itemID  string
		sizeKey string
		count   int
		wantErr error
	}{
		{subcat: "game_sets", itemID: "001", sizeKey: "OS", count: 1, wantErr: nil},                           // OK
		{subcat: "game_sets", itemID: "002", sizeKey: "OS", count: 2, wantErr: nil},                           // OK
		{subcat: "posters", itemID: "003", sizeKey: "OS", count: 2, wantErr: nil},                             // OK
		{subcat: "posters", itemID: "004", sizeKey: "OS", count: 2, wantErr: fmt.Errorf(ErrConditionalCheck)}, // OUT OF STOCK
		{subcat: "shirts", itemID: "005", sizeKey: "XL", count: 1, wantErr: nil},                              // multiple keys in entry
		{subcat: "posters", itemID: "006", sizeKey: "OS", count: 2, wantErr: fmt.Errorf(ErrConditionalCheck)}, // PARTITION DOES NOT EXIST
		{subcat: "shirts", itemID: "007", sizeKey: "OS", count: 2, wantErr: fmt.Errorf(ErrConditionalCheck)},  // ITEM DOES NOT EXIST
	}

	table := NewTable(StoreItemsTable, StoreItemPK, StoreItemSK)
	tables := []Table{table}
	dbInfo := InitDB(tables)

	for _, test := range tests {
		id, err := UpdateInventoryCount(dbInfo, test.subcat, test.itemID, test.sizeKey, test.count)
		t.Logf("out of stock ID: %s", id)
		if err != nil && test.wantErr == nil {
			t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
		}
		if err != nil && test.wantErr != nil {
			if err.Error() != test.wantErr.Error() {
				t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
			}
		}
	}
}
