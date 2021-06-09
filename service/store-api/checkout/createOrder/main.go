package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/timeops"
)

const route = "/add-to-cart" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

type customerInfo struct {
	UserID          string `json:"user_id"`
	UserEmail       string `json:"user_email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	ShippingAddress string `json:"shipping_address"`
	ShippingCity    string `json:"shipping_city"`
	ShippingState   string `json:"shipping_state"`
	ShippingCountry string `json:"shipping_country"`
	ShippingZip     string `json:"shipping_zip"`
}

type billingInfo struct {
	BillingName    string `json:"billing_name"`
	BillingAddress string `json:"billing_address"`
	BililngCity    string `json:"billing_city"`
	BillingState   string `json:"billing_state"`
	BillingCountry string `json:"billing_country"`
	BillingZip     string `json:"billing_zip"`
}

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // users table
		Name:       dbops.CustomersTable,
		PrimaryKey: dbops.CustomersPK,
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
	data := customerInfo{}
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

	// get customer
	cust, err := dbops.GetCustomer(DB, data.UserEmail)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// get user cart
	cart, err := dbops.GetShoppingCart(DB, cust.UserID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// create order
	order := createOrder(data, cust, cart)

	// put order
	err = dbops.PutOrder(DB, order)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// update customer record
	err = dbops.PutCustomer(DB, cust)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	httpops.ErrResponse(w, "Successfully retreived site info: ", successMsg, http.StatusOK)
	return
}

func createOrder(info customerInfo, cust *store.Customer, cart *store.ShoppingCart) *store.Order {
	// crate order & set intitial fields
	order := &store.Order{}
	order.OrderID = generateOrderID(cust.UserID, cust.Orders)
	order.UserID = cust.UserID

	for _, item := range cart.Items {
		order.Items = append(order.Items, item)
	}
	order.SalesSubtotal = cart.Subtotal
	order.TotalItems = cart.TotalItems

	addr := createAddressString(info.ShippingAddress, info.ShippingCity, info.ShippingState, info.ShippingCountry, info.ShippingZip)
	order.ShippingFirstName, order.ShippingLastName = info.FirstName, info.LastName
	order.ShippingAddress = addr

	order.TtlMs = 300000 // 5 minutes
	initTime := time.Now()
	initTimeStr := timeops.ConvertToTimestampString(initTime)
	orderDateStr := timeops.ConvertToDateString(initTime)
	order.InitTime = initTimeStr
	order.OrderDate = orderDateStr

	// update customer object
	cust.Orders += 1
	cust.OpenOrder = true

	return order
}

func generateOrderID(userID string, orderCt int) string {
	orderID := fmt.Sprintf("%s-%d", userID, orderCt+1)
	return orderID
}

func createAddressString(addr, state, city, country, zip string) string {
	fmt := fmt.Sprintf("%s, %s, %s, %s %s", addr, state, city, country, zip)
	return fmt
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
