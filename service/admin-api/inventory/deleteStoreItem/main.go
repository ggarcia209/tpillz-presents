package main

/* deleteStoreItem deletes a StoreItem and it's corresponding StoreItemSummary object */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/admin/inventory/delete_item" // DELETE
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
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// get query strings from GET call
	params := httpops.GetQueryStringParams(r)
	subcat := params["sub_category"]
	itemID := params["item_id"]

	// remove from index
	index, err := dbops.GetStoreItemIndex(DB, subcat)
	if err != nil {
		log.Printf("RootHandler failed - get index: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	index.Remove(itemID)

	err = dbops.PutStoreItemIndex(DB, index)
	if err != nil {
		log.Printf("RootHandler failed - put index: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// delete item summary
	err = dbops.DeleteStoreItemSummary(DB, subcat, itemID)
	if err != nil {
		log.Printf("RootHandler failed - delete summary: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// delete item
	err = dbops.DeleteStoreItem(DB, subcat, itemID)
	if err != nil {
		log.Printf("RootHandler failed - delete item: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Item deleted!", itemID, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
