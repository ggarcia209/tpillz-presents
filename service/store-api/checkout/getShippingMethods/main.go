package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/apex/gateway"
	"github.com/coldbrewcloud/go-shippo"
	"github.com/coldbrewcloud/go-shippo/client"
	"github.com/coldbrewcloud/go-shippo/models"
	"github.com/go-aws/go-dynamo/dynamo"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
	"github.com/tpillz-presents/service/util/sortops"
)

// UPDATE
const route = "/checkout/new-order" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// getShippingMethods retrieves the available shipping methods and calculates the
// price for each option before returning to user.

// shippo api ops

// customerInfo represents the form info submitted to the checkout page
// IN-PROGRESS - get shipping cost (shippo api)
type customerInfo struct {
	UserID       string  `json:"user_id"`
	UserEmail    string  `json:"user_email"`
	OrderID      string  `json:"order_id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Company      string  `json:"company"`
	AddressLine1 string  `json:"address_line_1"`
	AddressLine2 string  `json:"address_line_2"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	Zip          string  `json:"zip"`
	PhoneNumber  string  `json:"phone_number"`
	ShippingCost float32 `json:"shipping_cost"`
}

type dimensions struct {
	Weight    float32
	Volume    float32
	MaxLength float32
	MaxWidth  float32
	MaxHeight float32
}

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // users table
		Name:       dbops.CustomersTable,
		PrimaryKey: dbops.CustomersPK,
		SortKey:    ""},
	dbops.Table{ // transactions table
		Name:       dbops.OrdersTable,
		PrimaryKey: dbops.OrdersPK,
		SortKey:    dbops.OrdersSK},
}

// / DB is used to make DynamoDB API calls
var DB = dbops.InitDB(tables)

var shippoPrivateToken = "test"

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: return order summary first; get shipping info next
	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := customerInfo{}
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

	// initialize shippo client
	c := shippo.NewClient(shippoPrivateToken)

	// get order items
	order, err := dbops.GetOrderItems(DB, data.UserID, data.OrderID)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// get shipping rates
	rates, shipment, err := getShippingRates(DB, c, data, order)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// update order shipping address
	addr := createAddress(data)
	err = dbops.UpdateOrderAddress(DB, data.UserID, data.OrderID, addr, true)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), "SAVE_SHIPPING_ADDRESS_FAIL", http.StatusInternalServerError)
		return
	}

	// create shipment in DB
	err = dbops.PutShipment(DB, &shipment)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), "SAVE_SHIPPING_ADDRESS_FAIL", http.StatusInternalServerError)
		return
	}

	// return shipping rates
	httpops.ErrResponse(w, "Shipping rates: ", rates, http.StatusOK)
	return
}

func createAddress(info customerInfo) store.Address {
	addr := store.Address{
		FirstName:    info.FirstName,
		LastName:     info.LastName,
		Company:      info.Company,
		AddressLine1: info.AddressLine1,
		AddressLine2: info.AddressLine2,
		City:         info.City,
		State:        info.State,
		Country:      info.Country,
		Zip:          info.Zip,
		PhoneNumber:  info.PhoneNumber,
	}
	return addr
}

// get shipping rates for order
func getShippingRates(DB *dynamo.DbInfo, c *client.Client, data customerInfo, order *store.Order) ([]store.RateSummary, store.Shipment, error) {
	// create to/from addresses
	to, err := createShipmentAddress(c, data)
	if err != nil {
		log.Printf("createShipment failed: %v", err)
		return nil, store.Shipment{}, err
	}
	from, err := createReturnAddress(c)
	if err != nil {
		log.Printf("createShipment failed: %v", err)
		return nil, store.Shipment{}, err
	}
	// create parcels
	parcelObjs, err := dbops.GetParcels(DB, store.CarriersUsps)
	if err != nil {
		log.Printf("createShipment failed: %v", err)
		return nil, store.Shipment{}, err
	}

	parcels, packages, err := createParcels(c, order.Items, parcelObjs)
	if err != nil {
		log.Printf("createShipment failed: %v", err)
		return nil, store.Shipment{}, err
	}

	// create shipment objects and get rates
	shipmentInput := &models.ShipmentInput{
		AddressFrom: from,
		AddressTo:   to,
		Parcels:     parcels,
	}
	shipment, err := c.CreateShipment(shipmentInput)
	if err != nil {
		log.Printf("createShipment failed: %v", err)
		return nil, store.Shipment{}, err
	}
	// return object to store in DB for further actioning
	shipmentDB := createShipmentObject(data, shipment, packages)

	rates := getRates(shipment.Rates)

	return rates, shipmentDB, nil
}

// create shippo address object with customer info
func createShipmentAddress(c *client.Client, data customerInfo) (*models.Address, error) {
	ai := &models.AddressInput{
		Name:     data.FirstName + " " + data.LastName,
		Company:  data.Company,
		Street1:  data.AddressLine1,
		Street2:  data.AddressLine2,
		City:     data.City,
		Zip:      data.Zip,
		State:    data.State,
		Country:  data.Country,
		Phone:    data.PhoneNumber,
		Email:    data.UserEmail,
		Validate: true,
	}
	// populate other fields if applicable
	addr, err := c.CreateAddress(ai)
	if err != nil {
		log.Printf("createShipmentAddress failed: %v", err)
		return nil, err
	}
	log.Printf("validation result: %v; %v", addr.ValidationResults.IsValid, addr.ValidationResults.Messages)
	if !addr.ValidationResults.IsValid {
		return nil, fmt.Errorf("INVALID_ADDRESS")
	}
	return addr, nil
}

// create shippo address object with business info
func createReturnAddress(c *client.Client) (*models.Address, error) {
	data := store.ReturnAddress
	ai := &models.AddressInput{
		Name:     data.FirstName + " " + data.LastName,
		Company:  data.Company,
		Street1:  data.AddressLine1,
		Street2:  data.AddressLine2,
		City:     data.City,
		Zip:      data.Zip,
		State:    data.State,
		Country:  data.Country,
		Phone:    data.PhoneNumber,
		Email:    data.Email,
		Validate: false,
	}
	// populate other fields if applicable
	addr, err := c.CreateAddress(ai)
	if err != nil {
		log.Printf("createShipmentAddress failed: %v", err)
		return nil, err
	}

	return addr, nil
}

// Create parcel(s) for order. Uses greedy algorithm for large multi-parcel orders to fit as many
// objects into the largest parcel as possible (higher price : volume ratio) and fit the remainder in the smallest
// parcel as possible and repeats as necessary for orders requring >2 parcels.
func createParcels(c *client.Client, items []*store.CartItem, parcels []*store.Parcel) ([]*models.Parcel, []store.Package, error) {
	parcelObjs := []*models.Parcel{}
	packages := []store.Package{}
	resVolPct := float32(0.2)
	for {
		// get parcel dimension constraints from order
		dimensions, err := getDimensions(items)
		if err != nil {
			log.Printf("createParcels failed: %v", err)
			return parcelObjs, packages, err
		}
		totalWtLbs, totalVolume := dimensions.Weight, dimensions.Volume
		maxLength, maxWidth, maxHeight := dimensions.MaxLength, dimensions.MaxWidth, dimensions.MaxHeight

		// get parcel
		parcel, packaged, remVol, err := getParcelForVolume(c, parcels, totalWtLbs, totalVolume, maxLength, maxWidth, maxHeight, resVolPct)
		if err != nil {
			log.Printf("createParcels failed: %v", err)
			return parcelObjs, packages, err
		}
		// return if complete order packaged
		if remVol == 0.0 {
			parcelObjs = append(parcelObjs, parcel)
			packages = append(packages, packaged)
			return parcelObjs, packages, nil
		}

		// split packages with greedy algorithm if total order volume is greater than largest parcel volume
		sortedByVol := sortops.SortCartItemsByUnitVolume(items)
		pkg := []*store.CartItem{}  // items per individual package
		items = []*store.CartItem{} // remaining unpacked items

		// add items to parcel by greatest volume to least until full
		for _, item := range sortedByVol {
			if (item.ShippingDimensions.Volume * float32(item.Quantity)) <= remVol {
				// total volume of all items of same type fit in parcel
				// add item to parcel contents
				pkg = append(pkg, item)
				remVol -= (item.ShippingDimensions.Volume * float32(item.Quantity))
			} else {
				// total volume < parcel volume
				// add items by individual units until full
				items = append(items, item) // add item to list of remaining unpackaged items for next iteration
				if remVol < item.ShippingDimensions.Volume {
					continue
				}
				qt := item.Quantity
				single := item
				single.Quantity = 0
				for i := 0; i < qt; i++ {
					// add to parcel contents as singular units
					if item.ShippingDimensions.Volume <= remVol {
						single.Quantity++
						item.Quantity--
						remVol -= item.ShippingDimensions.Volume
						if item.Quantity == 0 {
							break
						}
					} else {
						break
					}
				}
				if single.Quantity > 0 {
					pkg = append(pkg, single)
				}
			}
		}

		// get parcel dimension constraints from package
		dimensions, err = getDimensions(pkg)
		if err != nil {
			log.Printf("createParcels failed: %v", err)
			return parcelObjs, packages, err
		}
		totalWtLbs, totalVolume = dimensions.Weight, dimensions.Volume
		maxLength, maxWidth, maxHeight = dimensions.MaxLength, dimensions.MaxWidth, dimensions.MaxHeight

		// get parcel
		parcel, packaged, remVol, err = getParcelForVolume(c, parcels, totalWtLbs, totalVolume, maxLength, maxWidth, maxHeight, resVolPct)
		if err != nil {
			log.Printf("createParcels failed: %v", err)
			return parcelObjs, packages, err
		}
		// add parcel to list if complete
		if remVol == 0.0 {
			parcelObjs = append(parcelObjs, parcel)
			packages = append(packages, packaged)
		} else {
			// return nil, error
		}
		if len(items) == 0 {
			break
		}
		if len(items) == len(sortedByVol) {
			// return err - no parcels found
			return parcelObjs, packages, fmt.Errorf("ERR_NO_PARCELS_FOUND")
		}

	}
	return parcelObjs, packages, nil

}

// get package dimensions required to fit order
func getDimensions(items []*store.CartItem) (dimensions, error) {
	totalWtLbs := float32(0.0)
	totalVolume := float32(0.0) // cubic inches
	maxLength := float32(0.0)
	maxWidth := float32(0.0)
	maxHeight := float32(0.0)

	// calculate order volume
	for _, item := range items {
		// get volume and weight
		floats, err := item.ShippingDimensions.GetFloats()
		if err != nil {
			log.Printf("getDimensions failed: %v", err)
			return dimensions{}, err
		}
		l, w, h, wt := floats[0], floats[1], floats[3], floats[4]
		volume := l * w * h
		totalWtLbs += (wt * float32(item.Quantity))
		totalVolume += volume

		// get max l, w, h
		if l > maxLength {
			maxLength = l
		}
		if w > maxWidth {
			maxWidth = w
		}
		if h > maxHeight {
			maxHeight = h
		}
	}

	dim := dimensions{totalWtLbs, totalVolume, maxLength, maxWidth, maxHeight}

	return dim, nil
}

// get smallest parcel for order volume
func getParcelForVolume(c *client.Client, parcels []*store.Parcel, weight, volume, ml, mw, mh, resPct float32) (*models.Parcel, store.Package, float32, error) {
	rem := float32(0.0)
	sorted := sortops.SortParcelsByVolume(parcels)
	for _, p := range sorted {
		rem = p.ParcelDimensions.Volume
		// get smallest package
		if volume < float32(p.ParcelDimensions.Volume*(1-resPct)) { // leave extra space for packaging materials
			// verify parcel dimensions fit largest items
			floats, err := p.ParcelDimensions.GetFloats()
			if err != nil {
				log.Printf("getParcelForVolume failed: %v", err)
				return &models.Parcel{}, store.Package{}, rem, nil
			}

			// compare dimensions of largest items to dimensions of parcel
			l, w, h, _ := floats[0], floats[1], floats[3], floats[4]
			if l < ml || w < mw || h < mh {
				// parcel does not fit largest objects
				continue
			}
			// create store.Package object for DB storage
			pkg := store.Package{
				Carrier:    p.Carrier,
				ParcelID:   p.ParcelID,
				Name:       p.Name,
				Dimensions: p.ParcelDimensions,
				Template:   p.Template,
			}
			// create shippo parcel object
			pi := &models.ParcelInput{
				Length:       p.ParcelDimensions.Length,
				Width:        p.ParcelDimensions.Width,
				Height:       p.ParcelDimensions.Height,
				DistanceUnit: p.ParcelDimensions.DistanceUnit,
				Weight:       fmt.Sprintf("%.2f", weight),
				MassUnit:     p.ParcelDimensions.MassUnit,
			}
			parcel, err := c.CreateParcel(pi)
			if err != nil {
				log.Printf("createParcels failed: %v", err)
				return &models.Parcel{}, store.Package{}, rem, nil
			}
			return parcel, pkg, 0.0, nil
		} else {
			continue
		}
	}

	// no parcel found - order volume > largest parcel volume
	return &models.Parcel{}, store.Package{}, rem, nil
}

func getRates(rates []*models.Rate) []store.RateSummary {
	summary := []store.RateSummary{}
	for _, rate := range rates {
		sl := store.ServiceLevel{
			Name:  rate.ServiceLevel.Name,
			Token: rate.ServiceLevel.Token,
			Terms: rate.ServiceLevel.Terms,
		}
		rs := store.RateSummary{
			Price:        rate.AmountLocal,
			Currency:     rate.Currency,
			Provider:     rate.Provider,
			Days:         rate.Days,
			ServiceLevel: sl,
		}
		summary = append(summary, rs)
	}
	return summary
}

// create store.Shipment object for order fullfillment
func createShipmentObject(user customerInfo, s *models.Shipment, pkgs []store.Package) store.Shipment {
	addr := store.Address{
		FirstName:    s.AddressTo.Name,
		Company:      s.AddressTo.Company,
		AddressLine1: s.AddressTo.Street1,
		AddressLine2: s.AddressTo.Street2,
		City:         s.AddressTo.City,
		State:        s.AddressTo.State,
		Country:      s.AddressTo.Country,
		Zip:          s.AddressTo.Zip,
		PhoneNumber:  s.AddressTo.Phone,
		Email:        s.AddressTo.Email,
	}

	rates := []store.RateSummary{}
	for _, rate := range s.Rates {
		sl := store.ServiceLevel{
			Name:  rate.ServiceLevel.Name,
			Token: rate.ServiceLevel.Token,
			Terms: rate.ServiceLevel.Terms,
		}
		rs := store.RateSummary{
			Price:        rate.AmountLocal,
			Currency:     rate.Currency,
			Provider:     rate.Provider,
			Days:         rate.Days,
			ServiceLevel: sl,
		}
		rates = append(rates, rs)
	}

	shipment := store.Shipment{
		UserID:      user.UserID,
		OrderID:     user.OrderID,
		Status:      s.Status,
		AddressTo:   addr,
		AddressFrom: store.ReturnAddress,
		Packages:    pkgs,
		Rates:       rates,
	}

	return shipment
}

func round(x float32) float32 {
	price := float64(x)
	unit := 0.01 // round float values to next cent
	return float32(math.Ceil(price/unit) * unit)
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
