package snsops

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-aws/go-sns/gosns"
	"github.com/tpillz-presents/service/store-api/store"
)

// const FulfillmentTopicARN = os.Getenv("fulfillmentTopicArn")

// InitSesh encapsulates the gosns.InitSesh() method and returns the SNS service
// as an interface{} type.
func InitSesh() interface{} {
	svc := gosns.InitSesh()
	return svc
}

// PublishOrderNotification pushlishes an Order to the Fulfillment topic for receipt processing
// and order fulfillment.
func PublishOrderNotification(svc interface{}, order *store.Order, fulfillmentTopicArn string) (string, error) {
	js, err := json.Marshal(order)
	if err != nil {
		log.Printf("PublishOrderNotification failed: %v", err)
		return "", err
	}
	msgStr := string(js)

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		msgID, err := gosns.Publish(svc, msgStr, fulfillmentTopicArn)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("PollStagingQueue failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("PollStagingQueue failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return msgID, nil
	}
}

// PublishShipmentUpdate publishes a *store.Shipment object to the Shipment topic. Topic subscribers update
// the Shipment object in the database and notify the customer of the shipment.
func PublishShipmentUpdate(svc interface{}, order *store.Shipment, shipmentTopicArn string) (string, error) {
	js, err := json.Marshal(order)
	if err != nil {
		log.Printf("PublishShipmentUpdate failed: %v", err)
		return "", err
	}
	msgStr := string(js)

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		msgID, err := gosns.Publish(svc, msgStr, shipmentTopicArn)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("PublishShipmentUpdate failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("PublishShipmentUpdate failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return msgID, nil
	}
}
