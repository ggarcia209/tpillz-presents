package s3ops

import (
	"log"
	"time"

	"github.com/go-aws/go-s3/gos3"
)

// SystemAssetsBucket contains the bucket name of the S3 bucket containing system assets
const SystemAssetsBucket = "tpillz-presents-dev-2"

// InitSesh encapsulates the goses.InitSesh() method and returns the SES service
// as an interface{} type.
func InitSesh() interface{} {
	svc := gos3.InitSesh()
	return svc
}

// GetReceiptHtmlTemplate retrieves the receipt email html template from
// the SystemAssetsBucket in S3 and returns it as a string.
func GetReceiptHtmlTemplate(svc interface{}) (string, error) {
	// generate receipt and email info
	key := "html/email-receipt-tmpl.html" // test only

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		obj, err := gos3.GetObject(svc, SystemAssetsBucket, key)
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

// GetOrderNotificationHtmlTemplate retrieves the order notification email html template from
// the SystemAssetsBucket in S3 and returns it as a string.
func GetOrderNotificationHtmlTemplate(svc interface{}) (string, error) {
	// generate receipt and email info
	key := "html/email-order-notification-tmpl.html" // test only

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		obj, err := gos3.GetObject(svc, SystemAssetsBucket, key)
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

// GetShippingNotificationHtmlTemplate retrieves the shipping notification email html template from
// the SystemAssetsBucket in S3 and returns it as a string.
func GetShippingNotificationHtmlTemplate(svc interface{}) (string, error) {
	// generate receipt and email info
	key := "html/email-shipping-notification-tmpl.html" // test only

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		obj, err := gos3.GetObject(svc, SystemAssetsBucket, key)
		if err != nil {
			if err.Error() == gos3.ErrNoSuchKey {
				log.Printf("GetShippingNotificationHtmlTemplate failed: %v", err)
				return "", err
			}
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("GetShippingNotificationHtmlTemplate failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("GetShippingNotificationHtmlTemplate failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return string(obj), nil
	}
}
