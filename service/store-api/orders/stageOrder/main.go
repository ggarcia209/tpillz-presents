package main

// stageOrder receives order Staging messages from the Staging Queue.
// Staging messages contain Order, Transaction, and Customer objects for
// in-progress orders. Staged orders are actioned once payment is successfully
// processed and a payment status confirmation message is received.

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/queueops"
)

const route = "/checkout/payment" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // customers table
		Name:       dbops.CustomersTable,
		PrimaryKey: dbops.CustomersPK,
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

	url, err := queueops.GetQueueURL(sqs, queueops.StagingFifoQueue)
	if err != nil {
		log.Printf("stageOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	// receive staged orders from queue
	resp, err := queueops.PollStagingQueue(sqs, url)
	if err != nil {
		log.Printf("stageOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Stages) == 0 {
		log.Printf("No stages to process - returning...")
		httpops.ErrResponse(w, "No stages to process.", successMsg, http.StatusOK)
		return
	}

	// put staged order info to database
	for _, stage := range resp.Stages {
		err := dbops.PutOrder(DB, stage.Order)
		if err != nil {
			log.Printf("stageOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		err = dbops.PutTransaction(DB, stage.Transaction)
		if err != nil {
			log.Printf("stageOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
		err = dbops.PutCustomer(DB, stage.Customer)
		if err != nil {
			log.Printf("stageOrder failed: %v", err)
			httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// delete stages from queue
	err = queueops.DeleteMessages(sqs, url, resp.MessageIDs, resp.ReceiptHandles)
	if err != nil {
		log.Printf("stageOrder failed: %v", err)
		httpops.ErrResponse(w, "Internal server error: ", err.Error(), http.StatusInternalServerError)
		return
	}

	httpops.ErrResponse(w, "Successfully staged order info: ", successMsg, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
