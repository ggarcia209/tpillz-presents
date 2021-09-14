package main

/* viewOpenOrders API gets open orders from the Fulfillment queue */

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/queueops"
)

const route = "/fulfillment/view-open-orders" // GET
const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	sqs := queueops.InitSesh()

	// poll queue for order
	url, err := queueops.GetQueueURL(sqs, queueops.FulfillmentFifoQueue)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// get order summary messages
	resp, err := queueops.PollFulfillmentQueue(sqs, url)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// return poll response to admin
	httpops.ErrResponse(w, "Order success! Receipt: : ", resp, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
