package main

/* purchaseLabel API purchases a shipping label  */

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/shipops"
	"github.com/tpillz-presents/service/util/snsops"
)

const route = "/fulfillment/purchase-label" // PUT
const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// shippo API private key - get as env var
const privateKey = ""

// AWS ARN for Shipping Topic - get as env var
const shipmentTopicArn = ""

// http request data
type request struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

// http response data
type responseBody struct {
	OrderID              string `json:"order_id"`
	Carrier              string `json:"carrier"`
	Price                string `json:"price"`
	LabelUrl             string `json:"label_url"`
	CommercialInvoiceUrl string `json:"commercial_invoice_url"`
	TrackingUrlProvider  string `json:"tracking_url_provider"`
	Eta                  string `json:"eta"`
}

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // orders table
		Name:       dbops.OpenOrdersTable,
		PrimaryKey: dbops.OpenOrdersPK,
		SortKey:    dbops.OpenOrdersSK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)
	sns := snsops.InitSesh()

	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := request{}
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			httpops.ErrResponse(w, "Bad Request: Wrong type provided for field "+unmarshalErr.Field, failMsg, http.StatusBadRequest)
		} else {
			httpops.ErrResponse(w, "Bad Request: "+err.Error(), failMsg, http.StatusBadRequest)
		}
		return
	}

	// get shipment
	shipment, err := dbops.GetShipment(DB, data.UserID, data.OrderID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// initialize shippo client and purchase label
	c := shipops.InitClient(privateKey)
	err = shipops.PurchaseShippingLabel(c, shipment)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	resp := responseBody{
		OrderID:              shipment.OrderID,
		Carrier:              shipment.Labels[0].Carrier,
		Price:                shipment.Labels[0].Price,
		LabelUrl:             shipment.Labels[0].LabelUrl,
		CommercialInvoiceUrl: shipment.Labels[0].CommercialInvoiceUrl,
		TrackingUrlProvider:  shipment.Labels[0].TrackingUrlProvider,
		Eta:                  shipment.Labels[0].Eta,
	}

	// send shipment to shipping update topic >>> update shipment db object, send customer email notification
	msgID, err := snsops.PublishShipmentUpdate(sns, shipment, shipmentTopicArn)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), resp, http.StatusInternalServerError)
		return
	}
	log.Printf("message sent to shipment topic: %s", msgID)

	// return order to admin
	httpops.ErrResponse(w, "Success! Returning order...", resp, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
