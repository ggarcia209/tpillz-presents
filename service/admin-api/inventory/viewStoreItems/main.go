package main

/* viewStoreItems returns a list of items for the given category, similar as if browsing as a customer. */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/admin/inventory/view_items" // GET
const failMsg = "Request failed!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{
		Name:       dbops.StoreItemsSummaryTable(),
		PrimaryKey: dbops.StoreItemSummaryPK,
		SortKey:    dbops.StoreItemSummarySK,
	},
	dbops.Table{
		Name:       dbops.StoreItemsIndexTable(),
		PrimaryKey: dbops.StoreItemsIndexPK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// get query strings from GET call
	params := httpops.GetQueryStringParams(r)
	subcat := params["sub_category"]

	tn := dbops.StoreItemsIndexTable()
	log.Println("tablename: ", tn)

	index, err := dbops.GetStoreItemIndex(DB, subcat)
	if err != nil {
		log.Printf("RootHandler failed - get index: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// get open order
	items, err := dbops.BatchGetStoreItemSummary(DB, subcat, index.ItemIDs)
	if err != nil {
		log.Printf("RootHandler failed - batch get: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Returning items...", items, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
