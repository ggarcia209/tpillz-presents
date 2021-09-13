package sesops

import (
	"testing"

	"github.com/tpillz-presents/service/store-api/store"
)

func TestSendCustomerReceipt(t *testing.T) {
	var tests = []*store.Order{
		&store.Order{
			OrderID:       "t0001",
			UserEmail:     "danielgarcia95367@gmail.com",
			SalesSubtotal: 19.99,
			SalesTax:      0.80,
			ShippingCost:  5.99,
			OrderTotal:    26.78,
			Shipment: store.Shipment{
				AddressTo: store.Address{
					FirstName:    "Daniel",
					LastName:     "Garcia",
					AddressLine1: "3250 Hollis St",
					AddressLine2: "Apt 319",
					City:         "Oakland",
					State:        "CA",
					Zip:          "94608",
					PhoneNumber:  "209-534-0739",
				},
			},
			Items: []*store.CartItem{
				&store.CartItem{Name: "PawnWars Chess Set", Quantity: 2},
				&store.CartItem{Name: "PawnWars Sniper Poster", Quantity: 1},
			},
		},
	}
	svc := InitSesh()
	from := "dg.dev.test510@gmail.com"
	for _, test := range tests {
		err := SendCustomerReceipt(svc, from, test)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
	}
}

func TestSendOrderNotification(t *testing.T) {
	var tests = []*store.Order{
		&store.Order{
			OrderID:       "t0001",
			UserEmail:     "danielgarcia95367@gmail.com",
			SalesSubtotal: 19.99,
			SalesTax:      0.80,
			ShippingCost:  5.99,
			OrderTotal:    26.78,
			Shipment: store.Shipment{
				AddressTo: store.Address{
					FirstName:    "Daniel",
					LastName:     "Garcia",
					AddressLine1: "3250 Hollis St",
					AddressLine2: "Apt 319",
					City:         "Oakland",
					State:        "CA",
					Zip:          "94608",
					PhoneNumber:  "209-534-0739",
				},
			},
			Items: []*store.CartItem{
				&store.CartItem{Name: "PawnWars Chess Set", Quantity: 2},
				&store.CartItem{Name: "PawnWars Sniper Poster", Quantity: 1},
			},
		},
	}
	svc := InitSesh()
	from := "dg.dev.test510@gmail.com"
	to := "danielgarcia95367@gmail.com"
	for _, test := range tests {
		err := SendOrderNotification(svc, from, to, test)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
	}
}
