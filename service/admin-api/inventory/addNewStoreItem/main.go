package main

/* getOpenOrder API gets an open order from the OpenOrders table and returns it to the admin. */

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/hashops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/admin/inventory/add_new_item" // PUT
const failMsg = "Request failed!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // orders table
		Name:       dbops.StoreItemsTable(),
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK,
	},
	dbops.Table{ // orders table
		Name:       dbops.StoreItemsSummaryTable(),
		PrimaryKey: dbops.StoreItemSummaryPK,
		SortKey:    dbops.StoreItemSummarySK,
	},
	dbops.Table{ // orders table
		Name:       dbops.StoreItemsIndexTable(),
		PrimaryKey: dbops.StoreItemsIndexPK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := &store.StoreItem{}
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			httpops.ErrResponse(w, "Bad Request: Wrong type provided for field "+unmarshalErr.Field, failMsg, http.StatusBadRequest)
		} else {
			httpops.ErrResponse(w, "Bad Request: "+err.Error(), failMsg, http.StatusBadRequest)
		}
		return
	}

	// generate itemID / root SKU (additional SKU's per size used for units sold and other sales metrics)
	generateRootSKU(data)

	summary := createSummary(data)

	// get put new store item to DB
	err = dbops.PutStoreItem(DB, data)
	if err != nil {
		log.Printf("RootHandler failed - put item: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	err = dbops.PutStoreItemSummary(DB, summary)
	if err != nil {
		log.Printf("RootHandler failed - put summary: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// update store item index
	index, err := dbops.GetStoreItemIndex(DB, data.Subcategory)
	if err != nil {
		log.Printf("RootHandler failed - get index: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	if index.Subcategory == "" {
		index.Subcategory = data.Subcategory
	}
	index.Push(data.ItemID)

	err = dbops.PutStoreItemIndex(DB, index)
	if err != nil {
		log.Printf("RootHandler failed - put index: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Item added!", data.ItemID, http.StatusOK)
	return
}

func generateRootSKU(item *store.StoreItem) {
	// generate abreviation/acronym of product name
	abrv := ""
	spl := strings.Split(item.Name, " ")
	for _, s := range spl {
		abrv += strings.Split(s, "")[0]
	}
	// generate 8 character hash id
	hash := hashops.GetMD5Hash64Bit(item.Name)
	// concatenate abreviation and hash id to create product root SKU / ID
	rootSku := abrv + "-" + hash
	item.ItemID = rootSku
}

func createSummary(item *store.StoreItem) *store.StoreItemSummary {
	sum := &store.StoreItemSummary{
		ItemID:      item.ItemID,
		Subcategory: item.Subcategory,
		Name:        item.Name,
		Price:       item.Price,
	}
	if len(item.ImageUrls) > 0 {
		sum.ThumbnailUrl = item.ImageUrls[0]
	}
	return sum
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
