package store

import (
	"testing"
)

func TestRound(t *testing.T) {
	var tests = []struct {
		input float32
		want  float32
	}{
		{1.2444, 1.25},
		{0.7222333, 0.73},
		{1.000000, 1},
		{0.000000, 0},
		{0.000000001, 0.01},
		{1.799999, 1.80},
		{25.25252525, 25.26},
		{100.19001, 100.2},
	}
	for _, test := range tests {
		rounded := round(test.input)
		if rounded != test.want {
			t.Errorf("FAIL: %f; want: %f", rounded, test.want)
		} else {
			t.Log(rounded)
		}
	}
}

func TestGetPriceOzs(t *testing.T) {
	var tests = []struct {
		rate   float32
		weight float32
		want   float32 // price
	}{
		{0.50, 4.00, 2.00},
		{1.00, 5.50, 5.50},
		{.10, 5.55, 0.56},
		{0.00, 20.00, 0.00},
	}
	for _, test := range tests {
		m := ShippingMethod{
			RateUSD:        test.rate,
			RateWeightUnit: "OZ",
		}
		price, err := m.GetPriceOzs(test.weight)
		if err != nil {
			t.Errorf("FAIL - err: %v", err)
		}
		if price != test.want {
			t.Errorf("FAIL: %f; want: %f", price, test.want)
		}
		if price == test.want {
			t.Log(price)
		}
	}
}

func TestGetPriceLbs(t *testing.T) {
	var tests = []struct {
		rate   float32
		weight float32
		want   float32 // price
	}{
		{0.50, 4.00, 2.00},
		{1.00, 5.50, 5.50},
		{.10, 5.55, 0.56},
		{0.00, 20.00, 0.00},
	}
	for _, test := range tests {
		m := ShippingMethod{
			RateUSD:        test.rate,
			RateWeightUnit: "LB",
		}
		price, err := m.GetPriceLbs(test.weight)
		if err != nil {
			t.Errorf("FAIL - err: %v", err)
		}
		if price != test.want {
			t.Errorf("FAIL: %f; want: %f", price, test.want)
		}
		if price == test.want {
			t.Log(price)
		}
	}
}

func TestGetPriceKg(t *testing.T) {
	var tests = []struct {
		rate   float32
		weight float32
		want   float32 // price
	}{
		{0.50, 4.00, 2.00},
		{1.00, 5.50, 5.50},
		{.10, 5.55, 0.56},
		{0.00, 20.00, 0.00},
	}
	for _, test := range tests {
		m := ShippingMethod{
			RateUSD:        test.rate,
			RateWeightUnit: "KG",
		}
		price, err := m.GetPriceKgs(test.weight)
		if err != nil {
			t.Errorf("FAIL - err: %v", err)
		}
		if price != test.want {
			t.Errorf("FAIL: %f; want: %f", price, test.want)
		}
		if price == test.want {
			t.Log(price)
		}
	}
}
