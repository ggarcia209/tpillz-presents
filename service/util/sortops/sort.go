package sortops

import (
	"sort"

	"github.com/tpillz-presents/service/store-api/store"
)

// SortedKV repesents a string/float pair representing an object's ID & relevant $ total.
// type map[float32]int is used by the shipping operations of the Store API.
type SortedFloatIntKV struct {
	Key   float32
	Value int
}

// SortedTotalsMap is a sorted list of SortedKV objects.
type SortedFloatIntTotalsMap []SortedFloatIntKV

func (s SortedFloatIntTotalsMap) Len() int           { return len(s) }
func (s SortedFloatIntTotalsMap) Less(i, j int) bool { return s[i].Key < s[j].Key }
func (s SortedFloatIntTotalsMap) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func SortIntFloatsLeastToGreatest(m map[float32]int) []SortedFloatIntKV {
	sorted := SortedFloatIntTotalsMap{}
	for k, v := range m {
		st := SortedFloatIntKV{Key: k, Value: v}
		sorted = append(sorted, st)
	}
	sort.Sort(sorted)
	return sorted
}

// SortedCartItems is a list of *store.CartItem objects sorted by shipping volume greatest to least.
type SortedCartItems []*store.CartItem

func (s SortedCartItems) Len() int { return len(s) }
func (s SortedCartItems) Less(i, j int) bool {
	return s[i].ShippingDimensions.Volume > s[j].ShippingDimensions.Volume
}
func (s SortedCartItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// SortCartItemsByUnitVolume sorts a list of cart items by volume least to greatest.
func SortCartItemsByUnitVolume(items []*store.CartItem) SortedCartItems {
	sorted := SortedCartItems{}
	for _, item := range items {
		sorted = append(sorted, item)
	}
	sort.Sort(sorted)
	return sorted
}

// SortedParcels is a list of *store.CartParcel objects sorted by volume least to greatest.
type SortedParcels []*store.Parcel

func (s SortedParcels) Len() int { return len(s) }
func (s SortedParcels) Less(i, j int) bool {
	return s[i].ParcelDimensions.Volume < s[j].ParcelDimensions.Volume
}
func (s SortedParcels) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// SortParcelsByVolume sorts a list of parcels by volume least to greatest
func SortParcelsByVolume(parcels []*store.Parcel) SortedParcels {
	sorted := SortedParcels{}
	for _, parcel := range parcels {
		sorted = append(sorted, parcel)
	}
	sort.Sort(sorted)
	return sorted
}
