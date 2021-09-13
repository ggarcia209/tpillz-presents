package main

import (
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/queueops"
	"github.com/tpillz-presents/service/util/snsops"
)

const route = "/checkout/payment" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// FulfillmentTopicARN is the SNS ARN of the Fulfillment Topic. This environment variable's
// value is set in the SAM template.yaml file.
var FulfillmentTopicARN = os.Getenv("fulfillmentTopicArn")

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
	sns := snsops.InitSesh()

	url, err := queueops.GetQueueURL(sqs, queueops.PaymentStatusFifoQueue)
	if err != nil {
		log.Printf("processOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	// poll payment status queue
	resp, err := queueops.PollPaymentStatusQueue(sqs, url)
	if err != nil {
		log.Printf("processOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	// update staged orders
	for _, status := range resp.Statuses {
		custID := status.CustomerID
		check := store.ValidPaymentStatus[status.TxStatus]
		if !check {
			log.Printf("processOrder failed: INVALID_TX_STATUS")
			httpops.ErrResponse(w, "Internal server error: ", "INVALID_TX_STATUS: "+status.TxStatus, http.StatusInternalServerError)
			return
		}
		err := dbops.UpdateOrderPaymentStatus(DB, custID, status.OrderID, status.TxStatus)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		err = dbops.UpdateTxPaymentStatus(DB, custID, status.TransactionID, status.TxStatus)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		err = dbops.UpdateTxPaymentMethod(DB, custID, status.TransactionID, status.PaymentMethod)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		err = dbops.UpdateTxPaymentID(DB, custID, status.TransactionID, status.PaymentTxID)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}

		// get complete order record
		order, err := dbops.GetOrder(DB, custID, status.OrderID)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}

		// forward order to SNS topics
		msgID, err := snsops.PublishOrderNotification(sns, order, FulfillmentTopicARN)
		if err != nil {
			log.Printf("processOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("SNS message sent: %v", msgID)
	}

	// delete processed messages
	err = queueops.DeleteMessages(sqs, url, resp.MessageIDs, resp.ReceiptHandles)
	if err != nil {
		log.Printf("processOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	httpops.ErrResponse(w, "Successfully retreived site info: ", successMsg, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
