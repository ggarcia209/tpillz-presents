package main

// createOrder generates a new order after receiving user input shipping information.
// Order total price is calculated after receiving user input for shipping option.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/timeops"
)

// UPDATE
const route = "/checkout/new-order" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

const CASalesTaxRate = .0725 // 7.25 % CA State sales tax rate
const FeesTotal = 0.00

// customerInfo represents the form info submitted to the checkout page
// IN-PROGRESS - get shipping cost (shippo api)
type customerInfo struct {
	UserID         string  `json:"user_id"`
	UserEmail      string  `json:"user_email"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Company        string  `json:"company"`
	AddressLine1   string  `json:"address_line_1"`
	AddressLine2   string  `json:"address_line_2"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Country        string  `json:"country"`
	Zip            string  `json:"zip"`
	ShippingMethod string  `json:"shipping_method"`
	PhoneNumber    string  `json:"phone_number"`
	ShippingCost   float32 `json:"shipping_cost"`
}

type orderSummary struct {
	Message    string            `json:"message"`
	Items      []*store.CartItem `json:"items"`
	TotalItems int               `json:"total_items"`
	Subtotal   float32           `json:"subtotal"`
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
		Name:       dbops.OrdersTable,
		PrimaryKey: dbops.OrdersPK,
		SortKey:    dbops.OrdersSK},
	dbops.Table{ // transactions table
		Name:       dbops.TransactionsTable,
		PrimaryKey: dbops.TransactionsPK,
		SortKey:    dbops.TransactionsSK},
}

// / DB is used to make DynamoDB API calls
var DB = dbops.InitDB(tables)

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {

	// NOTE: return order summary first; get shipping info next
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
	cust, err := dbops.GetCustomer(DB, data.UserEmail) // change to user_id
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
	order := createOrder(cust, cart)

	// put order
	err = dbops.PutOrder(DB, order)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// update customer record
	// TO DO: update fields only with expression
	err = dbops.PutCustomer(DB, cust)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	summary := orderSummary{Message: successMsg, Items: order.Items, TotalItems: order.TotalItems, Subtotal: order.SalesSubtotal}

	httpops.ErrResponse(w, "Successfully retreived site info: ", summary, http.StatusOK)
	return
}

func createOrder(cust *store.Customer, cart *store.ShoppingCart) *store.Order {
	// crate order & set intitial fields
	order := &store.Order{}
	order.OrderID = generateOrderID(cust.UserID, cust.Orders)
	order.UserID = cust.UserID

	for _, item := range cart.Items {
		order.Items = append(order.Items, item)
	}
	order.SalesSubtotal = round(cart.Subtotal)
	// order.ShippingCost = round(info.ShippingCost)
	order.SalesTax = round(order.SalesSubtotal * CASalesTaxRate)
	order.ChargesAndFees = round(FeesTotal)
	order.OrderTotal = order.SalesSubtotal + order.ShippingCost + order.SalesTax + order.ChargesAndFees
	order.TotalItems = cart.TotalItems

	// addr := createAddress(info)
	// order.ShippingAddress = addr

	order.OrderWeightOzs = cart.CartWeightOzs
	order.OrderWeightLbs = cart.CartWeightLbs
	order.OrderWeightKgs = cart.CartWeightKgs

	order.TtlMs = 600000 // 10 minutes
	initTime := time.Now()
	initTimeStr := timeops.ConvertToTimestampString(initTime)
	orderDateStr := timeops.ConvertToDateString(initTime)
	order.InitTime = initTimeStr
	order.OrderDate = orderDateStr

	order.OrderStatus = store.OrderStatusOpen

	// update customer object
	if !cust.OpenOrder {
		cust.OpenOrder = true
	}
	cust.Orders += 1
	cust.OpenOrderIDs = append(cust.OpenOrderIDs, order.OrderID)

	return order
}

func generateOrderID(userID string, orderCt int) string {
	orderID := fmt.Sprintf("%s-%d", userID, orderCt+1)
	return orderID
}

func createAddress(info customerInfo) store.Address {
	addr := store.Address{
		FirstName:    info.FirstName,
		LastName:     info.LastName,
		Company:      info.Company,
		AddressLine1: info.AddressLine1,
		AddressLine2: info.AddressLine2,
		City:         info.City,
		State:        info.State,
		Country:      info.Country,
		Zip:          info.Zip,
		PhoneNumber:  info.PhoneNumber,
	}
	return addr
}

func round(x float32) float32 {
	price := float64(x)
	unit := 0.01 // round float values to next cent
	return float32(math.Ceil(price/unit) * unit)
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
