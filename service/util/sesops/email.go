package sesops

import (
	"fmt"
	"log"
	"time"

	"github.com/go-aws/go-ses/goses"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/htmlops"
	"github.com/tpillz-presents/service/util/s3ops"
)

// const FulfillmentTopicARN = os.Getenv("fulfillmentTopicArn")

// InitSesh encapsulates the gosns.InitSesh() method and returns the SNS service
// as an interface{} type.
func InitSesh() interface{} {
	svc := goses.InitSesh()
	return svc
}

// SendCustomerReceipt sends an email receipt to the customer.
// 'from' specifies the SES verified sender email (ex: orders@store.com)
func SendCustomerReceipt(svc interface{}, from string, order *store.Order) error {
	// generate receipt and email info
	// receipt := order.NewReceipt()
	subject := fmt.Sprintf("Thank you from ACamoPrjct! (Order #%s)", order.OrderID)
	text := fmt.Sprintf("Order #%s received! Price: %0.2f | Estimated Shipping Days: %d", order.OrderID, order.OrderTotal, order.Shipment.EstimatedDays)
	tmpl, err := s3ops.GetReceiptHtmlTemplate(s3ops.InitSesh())
	if err != nil {
		log.Printf("SendCustomerReceipt failed: %v", err)
		return err
	}
	items := []htmlops.ItemSummary{}
	for _, item := range order.Items {
		is := htmlops.ItemSummary{
			Name:     item.Name,
			Quantity: item.Quantity,
		}
		items = append(items, is)
	}
	htmlInput := htmlops.ReceiptTemplateData{
		OrderID:    order.OrderID,
		Subtotal:   order.SalesSubtotal,
		Shipping:   order.ShippingCost,
		SalesTax:   order.SalesTax,
		OrderTotal: order.OrderTotal,
		FirstName:  order.Shipment.AddressTo.FirstName,
		LastName:   order.Shipment.AddressTo.LastName,
		Address1:   order.Shipment.AddressTo.AddressLine1,
		Address2:   order.Shipment.AddressTo.AddressLine2,
		City:       order.Shipment.AddressTo.City,
		State:      order.Shipment.AddressTo.State,
		Zip:        order.Shipment.AddressTo.Zip,
		Phone:      order.Shipment.AddressTo.PhoneNumber,
		Items:      items,
	}
	html, err := htmlops.CreateOrderReceiptHtml(tmpl, htmlInput)
	if err != nil {
		log.Printf("SendCustomerReceipt failed: %v", err)
		return err
	}

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		err := goses.SendEmail(svc, []string{order.UserEmail}, []string{}, from, subject, text, html)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendCustomerReceipt failed: %v -- max retries exceeded", err)
				return err
			}
			log.Printf("SendCustomerReceipt failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return nil
	}
}

// SendOrderNotification sends an order notification email intended for the business admin and/or fulfillment team.
// 'from' specifies the 'from' address (ex: orders@store.com), 'notifyEmail' specifies the 'to' address (ex: fulfillment@store.com).
func SendOrderNotification(svc interface{}, from, notifyEmail string, order *store.Order) error {
	// generate receipt and email info
	subject := fmt.Sprintf("New Order! (#%s)", order.OrderID)
	text := fmt.Sprintf("Order #%s received! Price: %0.2f", order.OrderID, order.OrderTotal)
	tmpl, err := s3ops.GetOrderNotificationHtmlTemplate(s3ops.InitSesh())
	if err != nil {
		log.Printf("SendCustomerReceipt failed: %v", err)
		return err
	}

	items := []htmlops.ItemSummary{}
	for _, item := range order.Items {
		is := htmlops.ItemSummary{
			Name:     item.Name,
			Quantity: item.Quantity,
		}
		items = append(items, is)
	}
	htmlInput := htmlops.ReceiptTemplateData{
		OrderID:    order.OrderID,
		Subtotal:   order.SalesSubtotal,
		Shipping:   order.ShippingCost,
		SalesTax:   order.SalesTax,
		OrderTotal: order.OrderTotal,
		FirstName:  order.Shipment.AddressTo.FirstName,
		LastName:   order.Shipment.AddressTo.LastName,
		Address1:   order.Shipment.AddressTo.AddressLine1,
		Address2:   order.Shipment.AddressTo.AddressLine2,
		City:       order.Shipment.AddressTo.City,
		State:      order.Shipment.AddressTo.State,
		Zip:        order.Shipment.AddressTo.Zip,
		Phone:      order.Shipment.AddressTo.PhoneNumber,
		Items:      items,
	}
	html, err := htmlops.CreateOrderReceiptHtml(tmpl, htmlInput)
	if err != nil {
		log.Printf("SendCustomerReceipt failed: %v", err)
		return err
	}

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		err := goses.SendEmail(svc, []string{notifyEmail}, []string{}, from, subject, text, html)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendCustomerReceipt failed: %v -- max retries exceeded", err)
				return err
			}
			log.Printf("SendCustomerReceipt failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		return nil
	}
}
