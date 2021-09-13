package sortops

import (
	"testing"

	"github.com/tpillz-presents/service/store-api/store"
)

func TestSortIntFloatsLeastToGreatest(t *testing.T) {
	var tests = []struct {
		input map[float32]int
		want  []float32
	}{
		{input: map[float32]int{0.0: 5, 1.0: 4, 2.0: 3, 3.0: 2, 4.0: 1}, want: []float32{0.0, 1.0, 2.0, 3.0, 4.0}},
		{input: map[float32]int{1.1: 5, 3.33: 4, 6.67: 3, 2.2: 2, 0.05: 1}, want: []float32{0.05, 1.1, 2.2, 3.33, 6.67}},
		{input: map[float32]int{}, want: []float32{}},
	}
	for _, test := range tests {
		sorted := SortIntFloatsLeastToGreatest(test.input)
		for i, n := range sorted {
			if n.Key != test.want[i] {
				t.Logf("sorted: %v", sorted)
				t.Errorf("FAIL: %v; want: %v", n.Key, test.want[i])
			}
		}
	}
}

func TestSortCartItemsByUnitVolume(t *testing.T) {
	var tests = []struct {
		input []*store.CartItem
		want  SortedCartItems
	}{
		{input: []*store.CartItem{
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
			&store.CartItem{ItemID: "003", ShippingDimensions: store.Dimensions{Volume: 7.77}},
		}, want: SortedCartItems{
			&store.CartItem{ItemID: "003", ShippingDimensions: store.Dimensions{Volume: 7.77}},
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}},
		{input: []*store.CartItem{
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}, want: SortedCartItems{
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}},
		{input: []*store.CartItem{
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}, want: SortedCartItems{
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}},
		{input: []*store.CartItem{
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
		}, want: SortedCartItems{
			&store.CartItem{ItemID: "002", ShippingDimensions: store.Dimensions{Volume: 5.55}},
			&store.CartItem{ItemID: "001", ShippingDimensions: store.Dimensions{Volume: 3.33}},
		}},
		{input: []*store.CartItem{}, want: SortedCartItems{}},
	}
	for _, test := range tests {
		sorted := SortCartItemsByUnitVolume(test.input)
		for i, item := range sorted {
			if item.ItemID != test.want[i].ItemID {
				t.Log("sorted:")
				for j, s := range sorted {
					t.Logf("%d: {%s: %f}", j, s.ItemID, s.ShippingDimensions.Volume)
				}
				t.Errorf("FAIL: %v; want: %v", item.ItemID, test.want[i].ItemID)
			}
		}
	}
}

func TestSortParcelsByVolume(t *testing.T) {
	var tests = []struct {
		input []*store.Parcel
		want  SortedParcels
	}{
		{input: []*store.Parcel{
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
			&store.Parcel{ParcelID: "003", ParcelDimensions: store.Dimensions{Volume: 7.77}},
		}, want: SortedParcels{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
			&store.Parcel{ParcelID: "003", ParcelDimensions: store.Dimensions{Volume: 7.77}},
		}},
		{input: []*store.Parcel{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
		}, want: SortedParcels{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
		}},
		{input: []*store.Parcel{
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
		}, want: SortedParcels{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
		}},
		{input: []*store.Parcel{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
		}, want: SortedParcels{
			&store.Parcel{ParcelID: "001", ParcelDimensions: store.Dimensions{Volume: 3.33}},
			&store.Parcel{ParcelID: "002", ParcelDimensions: store.Dimensions{Volume: 5.55}},
		}},
		{input: []*store.Parcel{}, want: SortedParcels{}},
	}
	for _, test := range tests {
		sorted := SortParcelsByVolume(test.input)
		for i, item := range sorted {
			if item.ParcelID != test.want[i].ParcelID {
				t.Log("sorted:")
				for j, s := range sorted {
					t.Logf("%d: {%s: %f}", j, s.ParcelID, s.ParcelDimensions.Volume)
				}
				t.Errorf("FAIL: %v; want: %v", item.ParcelID, test.want[i].ParcelID)
			}
		}
	}
}
