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
	"github.com/tpillz-presents/service/util/queueops"
)

const route = "/checkout/payment" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

const CASalesTaxRate = .0725  // 7.25 % CA State sales tax rate

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
	UserID          string `json:"user_id"`
	UserEmail       string `json:"user_email"`
	FirstName       string `json:"first_name"`
	LCompany      string `json:"company"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Zip          string `json:"zip"`
	PhoneNumber  string `json:"phone_number"`
	SaveInfo bool `json:"save_info"`
}

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // customers table
		Name:       dbops.CustomersTable,
		PrimaryKey: dbops.CustomersPK,
		SortKey:    ""
	},
	dbops.Table{ // store items table
		Name:       dbops.StoreItemsTable,
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK
	},
	dbops.Table{ // shopping carts table
		Name:       dbops.ShoppingCartsTable,
		PrimaryKey: dbops.ShoppingCartsPK,
		SortKey:    ""
	},
	dbops.Table{ // transactions table
		Name:       dbops.TransactionsTable,
		PrimaryKey: dbops.TransactionsPK,
		SortKey:    dbops.TransactionsSK
	},
	dbops.Table{ // orders table
		Name:       dbops.OrdersTable,
		PrimaryKey: dbops.OrdersPK,
		SortKey:    dbops.OrdersSK
	},
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

	// get order
	orderID := generateOrderID(cust.UserID, cust.Orders)
	order, err := dbops.GetOrder(DB, cust.UserID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// create transaction object
	tx := &store.Transaction{
		TransactionID: generateTxID()
		UserID: cust.UserID,
		OrderID: order.OrderID,
		Timestamp: timeops.ConvertToTimestampString(time.Now())
		Amount: 0.0,
	}

	// update order
	updateOrder(data, cust, tx, order)

	// stage objects for processing
	stage := staging{
		Order: order,
		Customer: cust,
		Transaction: tx,
	}

	// create stripe charge


	// put transaction
	err = dbops.PutTransaction(DB, tx)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// put order
	err = dbops.PutOrder(DB, order)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// send order to order processing service (record in db, send to admin)

	// update customer record
	err = dbops.PutCustomer(DB, cust)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// generate customer receipt

	// send receipt to customer email notification service

	httpops.ErrResponse(w, "Successfully retreived site info: ", successMsg, http.StatusOK)
	return
}


func generateOrderID(userID string, orderCt int) string {
	orderID := fmt.Sprintf("%s-%d", userID, orderCt)
	return orderID
}

func createAddressString(addr, city, state, country, zip string) store.Address {
	fmt := fmt.Sprintf("%s, %s, %s, %s %s", addr, city, state, country, zip)
	return fmt
}

func updateOrder(info billingInfo, cust *store.Customer, tx *store.Transaction, order *store.Order) {
	order.Complete = true

	order.TransactionID = tx.TransactionID
	order.TxTimestamp = tx.Timestamp

	order.SalesSubtotal = tx.SalesSubtotal
	order.ShippingCost = tx.ShippingCost
	order.SalesTax = tx.SalesTax
	order.ChargesAndFees = tx.ChargesAndFees
	order.OrderTotal = tx.TotalAmount

	address := store.Address{
		FirstName: info.FirstName,
		LastName: info.LastName,
		Company: info.Company,
		AddressLine1: info.AddressLine1,
		AddressLine2: info.AddressLine2,
		City: info.City,
		State: info.State,
		Country: info.Country,
		Zip: info.Zip,
		PhoneNumber: info.PhoneNumber
	}
	order.BillingAddress = address

	// update customer info
	if info.SaveInfo {
		cust.ShippingAddress = order.ShippingAddress
		cust.BillingAddress = order.BillingAddress
	}
	cust.Purchases += order.TotalItems
	cust.TotalSpent += tx.TotalAmount
	cust.OpenOrder = false
}

// add error handling for duplicate queue creation
func sendStagingMessage(stage staging) (string, error) {
	url, err := gosqs.GetQueueURL(SQS, types.SiteInfoFifoQueue)
	if err != nil {
		log.Printf("sendValidationMessage failed: %v", err)
		return "", err
	}
	// re-encode to JSON
	json, err := json.Marshal(valid)
	if err != nil {
		log.Printf("sendValidationMessage failed: %v", err)
		return "", err
	}

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  gosqs.GenerateDedupeID(url),
		MessageGroupId:          gosqs.GenerateDedupeID(url),
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	resp, err := gosqs.SendMessage(SQS, options)
	if err != nil {
		log.Printf("sendValidationMessage failed: %v", err)
		return "", err
	}
	return resp.MessageId, nil
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
