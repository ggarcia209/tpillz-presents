package main

import (
	"log"
	"testing"

	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
)

func TestRootHandler(t *testing.T) {
	var tests = []struct {
		item *store.CartItem
	}{
		{item: &store.CartItem{ // in stock
			UserID:      "user001",
			Subcategory: "game_sets",
			ItemID:      "001",
			SizeID:      "001-OS",
			Name:        "PawnWars Game Set",
			Size:        "OS",
			Quantity:    1,
			Price:       29.95,
		}},
		{item: &store.CartItem{
			UserID:      "user002",
			Subcategory: "shirts",
			ItemID:      "005",
			SizeID:      "005-M",
			Name:        "ACamoPrjct Logo T-Shirt",
			Size:        "M",
			Quantity:    1,
			Price:       22.95,
		}},
		{item: &store.CartItem{ // Insufficient stock
			UserID:      "user001",
			Subcategory: "shirts",
			ItemID:      "005",
			SizeID:      "005-L",
			Name:        "ACamoPrjct Logo T-Shirt",
			Size:        "L",
			Quantity:    3,
			Price:       22.95,
		}},
		{item: &store.CartItem{ // Non existent partition
			UserID:      "user002",
			Subcategory: "pants",
			ItemID:      "010",
			SizeID:      "010-32",
			Name:        "ACamoPrjct Baller Pants",
			Size:        "32",
			Quantity:    1,
			Price:       44.95,
		}},
	}
	for _, test := range tests {
		err := rootHandlerSim(test.item)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
	}
}

func rootHandlerSim(data *store.CartItem) error {
	// get user cart
	cart, err := dbops.GetShoppingCart(DB, data.UserID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		return err
	}

	if cart.UserID == "" {
		cart.UserID = data.UserID
	}

	// add item to cart
	if cart.Items == nil {
		cart.Items = make(map[string]*store.CartItem)
	}
	if cart.Items[data.SizeID] == nil {
		data.ItemSubtotal = data.Price * float32(data.Quantity)
		cart.Items[data.SizeID] = data
	} else {
		item := cart.Items[data.SizeID]
		item.Quantity += data.Quantity
		item.ItemSubtotal += (data.Price * float32(data.Quantity))
	}
	cart.TotalItems += data.Quantity
	cart.Subtotal += (float32(data.Quantity) * data.Price)

	// update cart db record
	err = dbops.PutShoppingCart(DB, cart)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		return err
	}

	return nil
}
