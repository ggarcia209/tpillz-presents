package main

/* getStoreItem retrieves a StoreItem and returns it to the admin. */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/admin/inventory/get_item" // GET
const failMsg = "Request failed!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // orders table
		Name:       dbops.StoreItemsTable(),
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// get query strings from GET call
	params := httpops.GetQueryStringParams(r)
	subcat := params["sub_category"]
	itemID := params["item_id"]

	// get item
	item, err := dbops.GetStoreItem(DB, subcat, itemID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Returning item...", item, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
