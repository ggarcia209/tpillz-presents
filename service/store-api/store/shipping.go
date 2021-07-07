package store

import (
	"fmt"
	"log"
	"math"
)

const InvalidWeightErr = "INVALID_WEIGHT_UNIT"

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
