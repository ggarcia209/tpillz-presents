package s3ops

import (
	"log"
	"time"

	"github.com/go-aws/go-s3/gos3"
)

// InitSesh encapsulates the gosns.InitSesh() method and returns the SNS service
// as an interface{} type.
func InitSesh() interface{} {
	svc := gos3.InitSesh()
	return svc
}

// PublishOrderNotification pushlishes an Order to the Fulfillment topic for receipt processing
// and order fulfillment.
func GetReceiptHtmlTemplate(svc interface{}) (string, error) {
	// generate receipt and email info
	bucket := "tpillz-presents-dev-2"     // test only
	key := "html/email-receipt-tmpl.html" // test only

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		obj, err := gos3.GetObject(svc, bucket, key)
		if err != nil {
			if err.Error() == gos3.ErrNoSuchKey {
				log.Printf("GetReceiptHtmlTemplate failed: %v", err)
				return "", err
			}
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("GetReceiptHtmlTemplate failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("GetReceiptHtmlTemplate failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return string(obj), nil
	}
}

// PublishOrderNotification pushlishes an Order to the Fulfillment topic for receipt processing
// and order fulfillment.
func GetOrderNotificationHtmlTemplate(svc interface{}) (string, error) {
	// generate receipt and email info
	bucket := "tpillz-presents-dev-2"                // test only
	key := "html/email-order-notification-tmpl.html" // test only

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		obj, err := gos3.GetObject(svc, bucket, key)
		if err != nil {
			if err.Error() == gos3.ErrNoSuchKey {
				log.Printf("GetOrderNotificationTemplate failed: %v", err)
				return "", err
			}
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("GetOrderNotificationTemplate failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("GetOrderNotificationTemplate failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return string(obj), nil
	}
}
