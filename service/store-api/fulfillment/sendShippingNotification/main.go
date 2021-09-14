package main

/* sendEmail is triggered when an Order message is published to the Fulfillment topic.
   This function emails a receipt containing summary info of the order to the customer,
   and emails a corresponding message to an admin-facing email address. */

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/sesops"
)

const from = "dg.dev.test510@gmail.com"                   // test only - move to admin settings db table in prod
const notificationAddress = "danielgarcia95367@gmail.com" // test only - move to admin settings db table in prod

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	svc := sesops.InitSesh()
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		// fmt.Printf("[%s %s] Message = %s \n", record.EventSource, snsRecord.Timestamp, snsRecord.Message)
		msg := snsRecord.Message
		ship := &store.Shipment{}

		// unmarshall json string
		err := json.Unmarshal([]byte(msg), ship)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

		// send email receipt to customer
		err = sesops.SendShippingNotification(svc, from, ship.AddressTo.Email, ship)
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
