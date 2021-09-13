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
	"github.com/tpillz-presents/service/util/queueops"
	"github.com/tpillz-presents/service/util/timeops"
)

const route = "/checkout/payment" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"
const orderTimeoutMsg = "Order expired! Please restart the checkout process and try again."

const CASalesTaxRate = .0725 // 7.25 % CA State sales tax rate

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
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
	SameAsShipping bool   `json:"same_as_shipping"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Company        string `json:"company"`
	AddressLine1   string `json:"address_line_1"`
	AddressLine2   string `json:"address_line_2"`
	City           string `json:"city"`
	State          string `json:"state"`
	Country        string `json:"country"`
	Zip            string `json:"zip"`
	PhoneNumber    string `json:"phone_number"`
	PaymentMethod  string `json:"payment_method"`
	PaymentToken   string `json:"payment_token"` // token used to authorize payment
	SaveInfo       bool   `json:"save_info"`
}

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // customers table
		Name:       dbops.CustomersTable,
		PrimaryKey: dbops.CustomersPK,
		SortKey:    "",
	},
	dbops.Table{ // store items table
		Name:       dbops.StoreItemsTable,
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK,
	},
	dbops.Table{ // shopping carts table
		Name:       dbops.ShoppingCartsTable,
		PrimaryKey: dbops.ShoppingCartsPK,
		SortKey:    "",
	},
	dbops.Table{ // transactions table
		Name:       dbops.TransactionsTable,
		PrimaryKey: dbops.TransactionsPK,
		SortKey:    dbops.TransactionsSK,
	},
	dbops.Table{ // orders table
		Name:       dbops.OrdersTable,
		PrimaryKey: dbops.OrdersPK,
		SortKey:    dbops.OrdersSK,
	},
}

// / DB is used to make DynamoDB API calls
var DB = dbops.InitDB(tables)

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	sqs := queueops.InitSesh()

	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := billingInfo{}
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
	order, err := dbops.GetOrder(DB, cust.UserID, orderID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// check if existing order expired
	init, err := timeops.ConvertStringToTimestamp(order.InitTime)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return error if expired
	ttl := time.Since(init)
	if int(ttl.Milliseconds()) > order.TtlMs {
		log.Printf("Order expired - payment not processed.")
		httpops.ErrResponse(w, "Request expired!", orderTimeoutMsg, http.StatusOK)
		return
	}

	// create transaction object
	tx := createTx(cust.UserID, order)

	// update order
	updateOrder(data, cust, tx, order)

	// stage objects for processing
	stage := queueops.Staging{
		Order:       order,
		Customer:    cust,
		Transaction: tx,
	}

	// send objects to staging queue
	url, err := queueops.GetQueueURL(sqs, queueops.StagingFifoQueue)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}
	msgID, err := queueops.SendStagingMessage(sqs, url, stage)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		log.Printf("staged order: %v", stage)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}
	log.Printf("staging message sent: %v", msgID)

	// verify item in stock
	out := []string{}
	rollback := []*store.CartItem{}
	stockOk := true
	for _, item := range order.Items {
		_, err := dbops.UpdateInventoryCount(DB, item.Subcategory, item.ItemID, item.Size, item.Quantity)
		if err != nil {
			if err.Error() == dbops.ErrConditionalCheck {
				stockOk = false
				out = append(out, item.Name)
				log.Printf("RootHandler: item %s out of stock", item.ItemID)
			}
			log.Printf("RootHandler failed: %v", err)
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
			break
		}
		rollback = append(rollback, item)
	}
	if !stockOk {
		// send rollback message
		update := queueops.InventoryUpdate{
			UserEmail: order.UserID,
			OrderID:   order.OrderID,
			Items:     rollback,
		}
		url, err := queueops.GetQueueURL(sqs, queueops.InventoryUpdateFifoQueue)
		if err != nil {
			log.Printf("RootHandler failed: %v", err)
			msg := "Request failed! Order not processed."
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), msg, http.StatusInternalServerError)
			return
		}
		msgId, err := queueops.SendInventoryUpdateMessage(sqs, url, update)
		if err != nil {
			msg := "Request failed! Order not processed."
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), msg, http.StatusInternalServerError)
			return
		}
		log.Printf("rollback msgId: %s", msgId)
		// return to user
		httpops.ErrResponse(w, "ITEM_OUT_OF_STOCK", out, http.StatusOK)
		return
	}

	// process payment
	// IN PROGRESS
	paymentOk := true
	// rollback db updates if payment fails
	if !paymentOk {
		// send rollback message
		update := queueops.InventoryUpdate{
			UserEmail: order.UserID,
			OrderID:   order.OrderID,
			Items:     order.Items,
		}
		url, err := queueops.GetQueueURL(sqs, queueops.InventoryUpdateFifoQueue)
		if err != nil {
			log.Printf("RootHandler failed: %v", err)
			msg := "Request failed! Order not processed."
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), msg, http.StatusInternalServerError)
			return
		}
		msgId, err := queueops.SendInventoryUpdateMessage(sqs, url, update)
		if err != nil {
			msg := "Request failed! Order not processed."
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), msg, http.StatusInternalServerError)
			return
		}
		log.Printf("rollback msgId: %s", msgId)
		/* for _, item := range order.Items {
			neg := item.Quantity * -1
			err := dbops.UpdateInventoryCount(DB, item.Subcategory, item.ItemID, item.Size, neg)
			if err != nil {
				if err.Error() == dbops.ErrConditionalCheck {
					log.Printf("RootHandler failed: item %s out of stock", item.ItemID)
					msg := fmt.Sprintf("ITEM_OUT_OF_STOCK-%s", item.ItemID
				}
				log.Printf("RootHandler failed: %v", err)
				httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
			}
		} */
	}

	// update transaction object with payment info on payment completion
	// IN PROGRESS
	updateTx(tx, order)

	// send payment confirmation message
	status := createPaymentStatus(cust, order, tx)
	url, err = queueops.GetQueueURL(sqs, queueops.PaymentStatusFifoQueue)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}
	msgID, err = queueops.SendPaymentStatusMessage(sqs, url, status)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}
	log.Printf("payment status message sent: %v", msgID)

	// generate customer receipt to return to user
	receipt := store.Receipt{}
	receipt.New(cust, order)

	httpops.ErrResponse(w, "Order success! Receipt: : ", receipt, http.StatusOK)
	return
}

func generateOrderID(userID string, orderCt int) string {
	orderID := fmt.Sprintf("%s-%d", userID, orderCt)
	return orderID
}

func createTx(custID string, order *store.Order) *store.Transaction {
	tx := &store.Transaction{
		TransactionID:  "",
		UserID:         custID,
		OrderID:        order.OrderID,
		Timestamp:      timeops.ConvertToTimestampString(time.Now()),
		SalesSubtotal:  order.SalesSubtotal,
		SalesTax:       order.SalesTax,
		ShippingCost:   order.ShippingCost,
		ChargesAndFees: order.ChargesAndFees,
		TotalAmount:    order.OrderTotal,
		PaymentStatus:  store.PaymentStatusInProgress,
	}
	tx.SetHashID()
	return tx
}

func updateOrder(info billingInfo, cust *store.Customer, tx *store.Transaction, order *store.Order) {
	// order.Paid = true
	order.OrderStatus = store.OrderStatusPaymentInProgress

	order.TransactionID = tx.TransactionID
	order.TxTimestamp = tx.Timestamp

	// set shipping address
	if info.SameAsShipping {
		order.BillingAddress = order.ShippingAddress
	} else {
		address := store.Address{
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
		order.BillingAddress = address
	}

	// update customer info
	if info.SaveInfo {
		cust.ShippingAddress = order.ShippingAddress
		cust.BillingAddress = order.BillingAddress
	}
	cust.Purchases += order.TotalItems
	// update following after payment confirmed
	// cust.TotalSpent += tx.TotalAmount
	// cust.OpenOrder = false
}

func createReceipt(cust *store.Customer, order *store.Order) store.Receipt {
	receipt := store.Receipt{
		UserID:          cust.UserID,
		OrderID:         order.OrderID,
		TransactionID:   order.TransactionID,
		UserEmail:       cust.Email,
		OrderSummary:    order.Items,
		SalesSubtotal:   order.SalesSubtotal,
		ShippingCost:    order.ShippingCost,
		SalesTax:        order.SalesTax,
		ChargesAndFees:  order.ChargesAndFees,
		OrderTotal:      order.OrderTotal,
		BillingAddress:  order.BillingAddress,
		ShippingAddress: order.ShippingAddress,
	}
	return receipt
}

func createPaymentStatus(cust *store.Customer, order *store.Order, tx *store.Transaction) store.PaymentStatus {
	status := store.PaymentStatus{
		CustomerEmail: cust.Email,
		CustomerID:    cust.UserID,
		OrderID:       order.OrderID,
		TransactionID: tx.TransactionID,
		PaymentMethod: tx.PaymentMethod,
		PaymentTxID:   tx.PaymentTxID,
		TxStatus:      tx.PaymentStatus,
		TxMessage:     tx.PaymentMessage,
	}
	return status
}

func updateTx(tx *store.Transaction, order *store.Order) {
	tx.PaymentMethod = "TEST"
	tx.PaymentTxID = "TEST0001"
	// payment tx timestamp?
	tx.PaymentStatus = store.PaymentStatusSuccess
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
