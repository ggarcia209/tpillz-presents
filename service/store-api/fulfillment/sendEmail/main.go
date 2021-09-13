package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-aws/go-ses/goses"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/sesops"
)

// UPDATE
const route = "/fulfillment/email" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

const from = "dg.dev.test510@gmail.com"                   // test only - move to admin settings db table in prod
const notificationAddress = "danielgarcia95367@gmail.com" // test only - move to admin settings db table in prod

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	svc := goses.InitSesh()
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

		// send email receipt to customer
		err = sesops.SendCustomerReceipt(svc, from, order)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

		// email receipt to admin
		err = sesops.SendOrderNotification(svc, from, notificationAddress, order)
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
