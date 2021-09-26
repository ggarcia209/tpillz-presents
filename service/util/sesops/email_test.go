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
			ShippingAddress: store.Address{
				FirstName:    "Daniel",
				LastName:     "Garcia",
				AddressLine1: "3250 Hollis St",
				AddressLine2: "Apt 319",
				City:         "Oakland",
				State:        "CA",
				Zip:          "94608",
				PhoneNumber:  "209-534-0739",
			},
			Items: []*store.CartItem{
				&store.CartItem{Name: "PawnWars Chess Set", Quantity: 2},
				&store.CartItem{Name: "PawnWars Sniper Poster", Quantity: 1},
			},
		},
		&store.Order{
			OrderID:       "t0002",
			UserEmail:     "sgarza1209@gmail.com",
			SalesSubtotal: 19.99,
			SalesTax:      0.80,
			ShippingCost:  5.99,
			OrderTotal:    26.78,
			ShippingAddress: store.Address{
				FirstName:    "Sal",
				LastName:     "Garza",
				AddressLine1: "3204 Caraway Ct",
				AddressLine2: "",
				City:         "Modesto",
				State:        "CA",
				Zip:          "95355",
				PhoneNumber:  "209-495-5130",
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
	var tests = []struct {
		order *store.Order
		to    string
	}{
		{order: &store.Order{
			OrderID:       "t0001",
			UserEmail:     "danielgarcia95367@gmail.com",
			SalesSubtotal: 19.99,
			SalesTax:      0.80,
			ShippingCost:  5.99,
			OrderTotal:    26.78,
			ShippingAddress: store.Address{
				FirstName:    "Daniel",
				LastName:     "Garcia",
				AddressLine1: "3250 Hollis St",
				AddressLine2: "Apt 319",
				City:         "Oakland",
				State:        "CA",
				Zip:          "94608",
				PhoneNumber:  "209-534-0739",
			},
			Items: []*store.CartItem{
				&store.CartItem{Name: "PawnWars Chess Set", Quantity: 2},
				&store.CartItem{Name: "PawnWars Sniper Poster", Quantity: 1},
			},
		}, to: "danielgarcia95367@gmail.com"},
		{order: &store.Order{
			OrderID:       "t0002",
			UserEmail:     "sgarza1209@gmail.com",
			SalesSubtotal: 19.99,
			SalesTax:      0.80,
			ShippingCost:  5.99,
			OrderTotal:    26.78,
			ShippingAddress: store.Address{
				FirstName:    "Sal",
				LastName:     "Garza",
				AddressLine1: "3204 Caraway Ct",
				AddressLine2: "",
				City:         "Modesto",
				State:        "CA",
				Zip:          "95355",
				PhoneNumber:  "209-495-5130",
			},
			Items: []*store.CartItem{
				&store.CartItem{Name: "PawnWars Chess Set", Quantity: 2},
				&store.CartItem{Name: "PawnWars Sniper Poster", Quantity: 1},
			},
		}, to: "sgarza1209@gmail.com"},
	}
	svc := InitSesh()
	from := "dg.dev.test510@gmail.com"
	for _, test := range tests {
		err := SendOrderNotification(svc, from, test.to, test.order)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
	}
}

func TestSendShippingNotification(t *testing.T) {
	var tests = []struct {
		shipment *store.Shipment
		to       string
	}{
		{shipment: &store.Shipment{
			UserID:  "u0001",
			OrderID: "o0001",
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
			Labels: []store.ShippingLabel{
				store.ShippingLabel{
					Carrier:             "USPS",
					TrackingNumber:      "tn123456789",
					TrackingUrlProvider: "https://acamoprjct.com",
					Eta:                 "9/19/2021",
				},
			},
			Packages: []store.Package{
				store.Package{},
			},
		}, to: "danielgarcia95367@gmail.com"},
		{shipment: &store.Shipment{
			UserID:  "u0002",
			OrderID: "o0002",
			AddressTo: store.Address{
				FirstName:    "Sal",
				LastName:     "Garza",
				AddressLine1: "3204 Caraway Ct",
				AddressLine2: "",
				City:         "Modesto",
				State:        "CA",
				Zip:          "95355",
				PhoneNumber:  "209-495-5130",
			},
			Labels: []store.ShippingLabel{
				store.ShippingLabel{
					Carrier:             "USPS",
					TrackingNumber:      "tn123456789",
					TrackingUrlProvider: "https://acamoprjct.com",
					Eta:                 "9/19/2021",
				},
			},
			Packages: []store.Package{
				store.Package{},
			},
		}, to: "sgarza1209@gmail.com"},
		{shipment: &store.Shipment{ // no tracking url
			UserID:  "u0001",
			OrderID: "o0001",
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
			Labels: []store.ShippingLabel{
				store.ShippingLabel{
					Carrier:        "USPS",
					TrackingNumber: "tn123456789",
					Eta:            "9/19/2021",
				},
			},
			Packages: []store.Package{
				store.Package{},
			},
		}, to: "danielgarcia95367@gmail.com"},
	}
	svc := InitSesh()
	from := "dg.dev.test510@gmail.com"
	for _, test := range tests {
		err := SendShippingNotification(svc, from, test.to, test.shipment)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
	}
}
