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
	text := fmt.Sprintf("Order #%s received! Price: %0.2f", order.OrderID, order.OrderTotal)
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
		FirstName:  order.ShippingAddress.FirstName,
		LastName:   order.ShippingAddress.LastName,
		Address1:   order.ShippingAddress.AddressLine1,
		Address2:   order.ShippingAddress.AddressLine2,
		City:       order.ShippingAddress.City,
		State:      order.ShippingAddress.State,
		Zip:        order.ShippingAddress.Zip,
		Phone:      order.ShippingAddress.PhoneNumber,
		Items:      items,
	}
	html, err := htmlops.CreateHtmlTemplate(tmpl, htmlInput)
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
		FirstName:  order.ShippingAddress.FirstName,
		LastName:   order.ShippingAddress.LastName,
		Address1:   order.ShippingAddress.AddressLine1,
		Address2:   order.ShippingAddress.AddressLine2,
		City:       order.ShippingAddress.City,
		State:      order.ShippingAddress.State,
		Zip:        order.ShippingAddress.Zip,
		Phone:      order.ShippingAddress.PhoneNumber,
		Items:      items,
	}
	html, err := htmlops.CreateHtmlTemplate(tmpl, htmlInput)
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

// SendShippingNotification sends an order notification email intended for the business admin and/or fulfillment team.
// 'from' specifies the 'from' address (ex: orders@store.com), 'notifyEmail' specifies the 'to' address (ex: fulfillment@store.com).
func SendShippingNotification(svc interface{}, from, to string, shipment *store.Shipment) error {
	// generate receipt and email info
	subject := fmt.Sprintf("Order #%s Shipped!", shipment.OrderID)
	text := fmt.Sprintf("Order #%s shipped!", shipment.OrderID)
	tmpl, err := s3ops.GetShippingNotificationHtmlTemplate(s3ops.InitSesh())
	if err != nil {
		log.Printf("SendShippingNotification failed: %v", err)
		return err
	}

	htmlInput := htmlops.ShippingNotificationTemplateData{
		OrderID:        shipment.OrderID,
		Carrier:        shipment.Labels[0].Carrier,
		ParcelQty:      len(shipment.Packages),
		TrackingNumber: shipment.Labels[0].TrackingNumber,
		TrackingUrl:    shipment.Labels[0].TrackingUrlProvider,
		Eta:            shipment.Labels[0].Eta,
		FirstName:      shipment.AddressTo.FirstName,
		LastName:       shipment.AddressTo.LastName,
		Address1:       shipment.AddressTo.AddressLine1,
		Address2:       shipment.AddressTo.AddressLine2,
		City:           shipment.AddressTo.City,
		State:          shipment.AddressTo.State,
		Zip:            shipment.AddressTo.Zip,
		Phone:          shipment.AddressTo.PhoneNumber,
	}
	html, err := htmlops.CreateHtmlTemplate(tmpl, htmlInput)
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
		err := goses.SendEmail(svc, []string{to}, []string{}, from, subject, text, html)
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
