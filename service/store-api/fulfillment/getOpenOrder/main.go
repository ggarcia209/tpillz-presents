package main

/* getOpenOrder API gets an open order from the OpenOrders table and returns it to the admin. */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

const route = "/fulfillment/get-open-order" // GET
const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // orders table
		Name:       dbops.OpenOrdersTable,
		PrimaryKey: dbops.OpenOrdersPK,
		SortKey:    dbops.OpenOrdersSK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// get query strings from GET call
	params := httpops.GetQueryStringParams(r)
	userID := params["user_id"]
	orderID := params["order_id"]

	// get open order
	order, err := dbops.GetOpenOrder(DB, userID, orderID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Returning order...", order, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
