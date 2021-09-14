package store

import (
	"fmt"
	"log"
	"math"
)

const InvalidWeightErr = "INVALID_WEIGHT_UNIT"

// Shippo parcel template codes
const (
	UspsLargeFlatRate   = "USPS_LargeFlatRateBox"      // 12.25 x 12.25 x 6.00 in
	UspsMediumFlatRate1 = "USPS_MediumFlatRateBox1"    // 11.25 x 8.75 x 6.00 in
	UspsMediumFlatRate2 = "USPS_MediumFlatRateBox2"    // 14.00 x 12.00 x 3.50 in
	UspsRegionalRateA1  = "USPS_RegionalRateBoxA1"     // 10.13 x 7.13 x 5.00 in
	UspsRegionalRateA2  = "USPS_RegionalRateBoxA2"     // 13.06 x 11.06 x 2.50 in
	UspsRegionalRateB1  = "USPS_RegionalRateBoxB1"     // 12.25 x 10.50 x 5.50 in
	UspsRegionalRateB2  = "USPS_RegionalRateBoxB2"     // 16.25 x 14.50 x 3.00 in
	UspsSmallFlatRate1  = "USPS_SmallFlatRateBox"      // 8.69 x 5.44 x 1.75 in
	UspsSmallFlatRate2  = "USPS_SmallFlatRateEnvelope" // 10.00 x 6.00 x 4.00 in
)

const CarriersUsps = "USPS"
const CarriersDHL = "DHL"
const CarriersUPS = "UPS"
const CarriersFedEx = "FedEx"

// UspsSmallParcel represents a 6x6x6 box used for shipping through USPS.
/* var UspsLargeFlatRateParcel = Parcel{
	Length:       "12.25",
	Width:        "12.25",
	Height:       "6.00",
	DistanceUnit: "in",
	Weight:       "1.5",
	MassUnit:     "lbs",
	Volume:       900.375,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

var UspsMediumFlatRate1Parcel = Parcel{
	Length:       "11.25",
	Width:        "8.75",
	Height:       "6.00",
	DistanceUnit: "in",
	Weight:       "1.5",
	MassUnit:     "lbs",
	Volume:       590.625,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

var UspsMediumFlatRate2Parcel = Parcel{
	Length:       "14.00",
	Width:        "12.00",
	Height:       "3.50",
	DistanceUnit: "in",
	Weight:       "1.5",
	MassUnit:     "lbs",
	Volume:       588,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

var UspsRegionalRateA1Parcel = Parcel{
	Length:       "10.13",
	Width:        "7.13",
	Height:       "5.00",
	DistanceUnit: "in",
	Weight:       "1.5",
	MassUnit:     "lbs",
	Volume:       361.135,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

var UspsRegionalRateA2Parcel = Parcel{
	Length:       "10.13",
	Width:        "7.13",
	Height:       "5.00",
	DistanceUnit: "in",
	Weight:       "1.5",
	MassUnit:     "lbs",
	Volume:       361.135,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

// UspsSmallParcel represents a 6x6x6 box used for shipping through USPS.
var UspsSmallParcel = Parcel{
	Length:       "6",
	Width:        "6",
	Height:       "6",
	DistanceUnit: "in",
	Weight:       "1.1",
	MassUnit:     "lbs",
	Volume:       216.00,
	VolumeUnit:   "cubic_inches",
	Carrier:      "USPS",
}

*/

type ShippingMethod struct {
	MethodID          string  `json:"method_id"`
	Name              string  `json:"name"`
	RateUSD           float32 `json:"rate_usd"`            // shipping rate in US Dollars ($/weight unit)
	RateWeightUnit    string  `json:"rate_weight_unit"`    // weight unit required for calculating shipping price
	WeightInputOz     float32 `json:"weight_input_oz"`     // User input weight value in Oz for price calculation.
	WeightInputLbs    float32 `json:"weight_input_lbs"`    // weigt value in lbs.
	WeightInputKg     float32 `json:"weight_input_kg"`     // weight input in kilograms
	PriceOutputUSD    float32 `json:"price_output_usd"`    // calculated price value in US Dollars
	PriceOutputString string  `json:"price_output_string"` // get price as string value
}

// Parcel represents a shipping parcel type for use with the Shippo API.
type Parcel struct {
	Carrier          string     `json:"carrier"`   // carrier - DB PK
	ParcelID         string     `json:"parcel_id"` // DB SK
	Name             string     `json:"name"`
	ParcelDimensions Dimensions `json:"parcel_dimensions"`
	Template         string     `json:"template"`        // shippo parcel template
	UnitPriceUSD     float32    `json:"unit_price_usd"`  // price per unit
	UnitsAvailable   int        `json:"units_available"` // units in stock ready for shipping use
}

// Package represents a filled parcel in a Shipment.
type Package struct {
	Carrier        string           `json:"carrier"`   // carrier - DB PK
	ParcelID       string           `json:"parcel_id"` // DB SK
	Name           string           `json:"name"`
	Dimensions     Dimensions       `json:"parcel_dimensions"`
	Template       string           `json:"template"` // shippo parcel template
	Items          []PkgItemSummary `json:"items"`
	TrackingNumber string           `json:"tracking_number"`
}

// PkgItemSummary contains summary information for each item in the package.
type PkgItemSummary struct {
	Subcategory string `json:"subcategory"`
	ItemID      string `json:"item_id"`
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
}

// RateSummary contains summary information for an order's shipping rates
type RateSummary struct {
	Price        string       `json:"string"`
	Currency     string       `json:"currency"`
	Provider     string       `json:"provider"`
	Days         int          `json:"days"`
	ServiceLevel ServiceLevel `json:"service_level"`
}

// ServiceLevel contains info for the service level of a shipping option.
// (USPS Priority, Fedex Ground, etc...)
type ServiceLevel struct {
	Name  string `json:"name"`
	Token string `json:"token"`
	Terms string `json:"terms"`
}

// ShippingLabel contains info for a shipping label purchase.
type ShippingLabel struct {
	OrderID              string `json:"order_id"` // pk
	LabelID              string `json:"label_id"` // sk
	Carrier              string `json:"carrier"`
	Price                string `json:"price"`
	Currency             string `json:"currency"`
	PurchaseDate         string `json:"purchase_date"`
	TrackingNumber       string `json:"tracking_number"`
	TrackingStatus       string `json:"tracking_status"`
	TrackingUrlProvider  string `json:"tracking_url_provider"`
	Eta                  string `json:"eta"`
	LabelUrl             string `json:"label_url"`
	CommercialInvoiceUrl string `json:"commercial_invoice_url"`
}

// Shipment contains order shipping info used for order fulfillment at the time of shipping label purchase.
type Shipment struct {
	UserID        string          `json:"user_id"`  // pk
	OrderID       string          `json:"order_id"` // sk
	Status        string          `json:"status"`
	AddressTo     Address         `json:"address_to"`
	AddressFrom   Address         `json:"address_from"`
	Packages      []Package       `json:"packages"`
	Rates         []RateSummary   `json:"rates"`
	SelectedRate  RateSummary     `json:"selected_rate"`
	Labels        []ShippingLabel `json:"labels"`
	EstimatedDays int             `json:"estimated_days"`
}

func (s *ShippingMethod) GetPriceOzs(weight float32) (float32, error) {
	if s.RateWeightUnit != "OZ" {
		log.Printf("invvalid weight unit %s for GetPricesOz()", s.RateWeightUnit)
		return 0.00, fmt.Errorf(InvalidWeightErr)
	}
	s.WeightInputOz = weight
	price := s.RateUSD * s.WeightInputOz
	rounded := round(price)
	s.PriceOutputUSD = rounded
	return rounded, nil
}

func (s *ShippingMethod) GetPriceLbs(weight float32) (float32, error) {
	if s.RateWeightUnit != "LB" {
		log.Printf("invvalid weight unit %s for GetPricesLbs()", s.RateWeightUnit)
		return 0.00, fmt.Errorf(InvalidWeightErr)
	}
	s.WeightInputLbs = weight
	price := s.RateUSD * s.WeightInputLbs
	rounded := round(price)
	s.PriceOutputUSD = rounded
	return rounded, nil
}

func (s *ShippingMethod) GetPriceKgs(weight float32) (float32, error) {
	if s.RateWeightUnit != "KG" {
		log.Printf("invvalid weight unit %s for GetPricesLbs()", s.RateWeightUnit)
		return 0.00, fmt.Errorf(InvalidWeightErr)
	}
	s.WeightInputKg = weight
	price := s.RateUSD * s.WeightInputKg
	rounded := round(price)
	s.PriceOutputUSD = rounded
	return rounded, nil
}

func round(x float32) float32 {
	price := float64(x)
	unit := 0.01 // round float values to next cent
	return float32(math.Ceil(price/unit) * unit)
}

func (s *ShippingMethod) GetPriceString() string {
	return fmt.Sprintf("%.2f", s.PriceOutputUSD)
}
