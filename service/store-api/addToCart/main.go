package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/dbops"
	"github.com/tpillz-presents/service/httpops"
	"github.com/tpillz-presents/service/store-api/store"
)

const route = "/add-to-cart" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // users table
		Name:       dbops.UsersTable,
		PrimaryKey: dbops.UsersPK,
		SortKey:    ""},
	dbops.Table{ // store items table
		Name:       dbops.StoreItemsTable,
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK},
	dbops.Table{ // shopping carts table
		Name:       dbops.ShoppingCartsTable,
		PrimaryKey: dbops.ShoppingCartsPK,
		SortKey:    ""},
	dbops.Table{ // transactions table
		Name:       dbops.TransactionsTable,
		PrimaryKey: dbops.TransactionsPK,
		SortKey:    dbops.TransactionsSK},
}

// / DB is used to make DynamoDB API calls
var DB = dbops.InitDB(tables)

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := store.CartItem{}
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

	// get user cart
	cart, err := dbops.GetShoppingCart(DB, data.UserID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// add item to cart
	if cart.Items == nil {
		cart.Items = make(map[string]*store.CartItem)
	}
	if cart.Items[data.SizeID] == nil {
		cart.Items[data.SizeID] = &data
	} else {
		item := cart.Items[data.SizeID]
		item.Quantity += data.Quantity
		item.ItemSubtotal += data.ItemSubtotal
	}

	// update cart db record
	err = dbops.PutShoppingCart(DB, cart)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	httpops.ErrResponse(w, "Successfully retreived site info: ", successMsg, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
