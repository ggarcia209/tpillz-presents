package main

/* getOpenOrder API gets an open order from the OpenOrders table and returns it to the admin. */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/store/browse" // GET
const failMsg = "Request failed!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // orders table
		Name:       dbops.StoreItemsSummaryTable,
		PrimaryKey: dbops.StoreItemSummaryPK,
		SortKey:    dbops.StoreItemSummarySK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// get query strings from GET call
	params := httpops.GetQueryStringParams(r)
	subcat := params["subcategory"]

	// get open order
	items, err := dbops.ScanItemsForCategory(DB, subcat)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
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
