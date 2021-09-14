package htmlops

import (
	"bytes"
	"html/template"
)

type ReceiptTemplateData struct {
	OrderID    string
	Subtotal   float32
	SalesTax   float32
	Shipping   float32
	OrderTotal float32
	FirstName  string
	LastName   string
	Address1   string
	Address2   string
	City       string
	State      string
	Zip        string
	Phone      string
	Items      []ItemSummary
}

type ItemSummary struct {
	Name     string
	Quantity int
}

type ShippingNotificationTemplateData struct {
	OrderID        string
	Carrier        string
	ParcelQty      int
	TrackingNumber string
	TrackingUrl    string
	Eta            string
	FirstName      string
	LastName       string
	Address1       string
	Address2       string
	City           string
	State          string
	Zip            string
	Phone          string
	Items          []ItemSummary
}

func CreateHtmlTemplate(tmpl string, data interface{}) (string, error) {
	t := template.New("order_notification")

	var err error
	t, err = t.Parse(tmpl)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	result := tpl.String()
	return result, nil
}
