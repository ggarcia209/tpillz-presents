package main

/* queueOrder is triggered by an SNS event when a new Order message is published to
the Fulfillment topic. This function writes the new order to the OpenOrders DB table
and sends a summary of the order to the Fulfillment queue. */

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/queueops"
)

// UPDATE
const route = "/fulfillment/email" // PUT

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

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	svc := queueops.InitSesh() // sqs
	db := dbops.InitDB(tables) // dynamodb

	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		// fmt.Printf("[%s %s] Message = %s \n", record.EventSource, snsRecord.Timestamp, snsRecord.Message)
		msg := snsRecord.Message
		order := &store.Order{}

		// unmarshall json string
		err := json.Unmarshal([]byte(msg), order)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

		// write order to open orders table
		err = dbops.PutOpenOrder(db, order)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

		// send order to fulfillment queue
		os := order.NewSummary()
		url, err := queueops.GetQueueURL(svc, queueops.FulfillmentFifoQueue)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}
		_, err = queueops.SendFulfillmentMessage(svc, url, os)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

	}
	return
}

func main() {
	lambda.Start(handler)
}
