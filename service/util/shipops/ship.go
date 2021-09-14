package shipops

import (
	"log"
	"time"

	"github.com/coldbrewcloud/go-shippo"
	"github.com/coldbrewcloud/go-shippo/client"
	"github.com/coldbrewcloud/go-shippo/models"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/timeops"
)

// InitClient initializes the Shippo API client.
func InitClient(privateKey string) *client.Client {
	c := shippo.NewClient(privateKey)
	return c
}

// CreateShipment creates a Shippo Shipment object from a *store.Shipment object.
func CreateShipment(c *client.Client, s *store.Shipment) (*models.Shipment, error) {
	// create a sending address
	a := s.AddressFrom
	addressFromInput := &models.AddressInput{
		Name:    a.FirstName + " " + a.LastName,
		Street1: a.AddressLine1,
		Street2: a.AddressLine2,
		Company: a.Company,
		City:    a.City,
		State:   a.State,
		Zip:     a.Zip,
		Country: a.Country,
		Phone:   a.PhoneNumber,
		Email:   a.Email,
	}
	addressFrom, err := c.CreateAddress(addressFromInput)
	if err != nil {
		log.Printf("CreateShipment failed: %v", err)
		return &models.Shipment{}, err
	}

	// create a receiving address
	a = s.AddressTo
	addressToInput := &models.AddressInput{
		Name:    a.FirstName + " " + a.LastName,
		Street1: a.AddressLine1,
		Street2: a.AddressLine2,
		Company: a.Company,
		City:    a.City,
		State:   a.State,
		Zip:     a.Zip,
		Country: a.Country,
		Phone:   a.PhoneNumber,
		Email:   a.Email,
	}
	addressTo, err := c.CreateAddress(addressToInput)
	if err != nil {
		log.Printf("CreateShipment failed: %v", err)
		return &models.Shipment{}, err
	}

	parcels := []*models.Parcel{}
	for _, pkg := range s.Packages {
		// create a parcel
		d := pkg.Dimensions
		parcelInput := &models.ParcelInput{
			Length:       d.Length,
			Width:        d.Width,
			Height:       d.Height,
			DistanceUnit: d.DistanceUnit,
			Weight:       d.Weight,
			MassUnit:     d.MassUnit,
			Template:     pkg.Template,
		}
		parcel, err := c.CreateParcel(parcelInput)
		if err != nil {
			log.Printf("CreateShipment failed: %v", err)
			return &models.Shipment{}, err
		}
		parcels = append(parcels, parcel)
	}

	// create a shipment
	shipmentInput := &models.ShipmentInput{
		AddressFrom: addressFrom,
		AddressTo:   addressTo,
		Parcels:     parcels,
		Async:       false,
	}
	shipment, err := c.CreateShipment(shipmentInput)
	if err != nil {
		log.Printf("CreateShipment failed: %v", err)
		return &models.Shipment{}, err
	}

	return shipment, nil
}

// PurchaseShippingLabel purchases a new shipping label for the given shipment object.
// Label is purchased per the Shipment's 'SelectedRate' field.
func PurchaseShippingLabel(c *client.Client, s *store.Shipment) error {
	// create shippo shipment object
	shipment, err := CreateShipment(c, s)
	if err != nil {
		log.Printf("PurchaseShippingLabel failed: %v", err)
		return err
	}

	// get rate
	for _, rate := range shipment.Rates {
		if rate.ServiceLevel.Token == s.SelectedRate.ServiceLevel.Token {
			shipment.Rates = []*models.Rate{rate}
			break
		}
	}

	// purchase label
	transactionInput := &models.TransactionInput{
		Rate:          shipment.Rates[0].ObjectID,
		LabelFileType: models.LabelFileTypePDF,
		Async:         false,
	}
	transaction, err := c.PurchaseShippingLabel(transactionInput)
	if err != nil {
		log.Printf("PurchaseShippingLabel failed: %v", err)
		return err
	}

	// update store.Shipment object
	label := store.ShippingLabel{
		OrderID:              s.OrderID,
		LabelID:              transaction.ObjectID,
		Carrier:              s.SelectedRate.Provider,
		Price:                s.SelectedRate.Price,
		Currency:             s.SelectedRate.Currency,
		PurchaseDate:         timeops.ConvertToDateString(time.Now()),
		TrackingNumber:       transaction.TrackingNumber,
		TrackingStatus:       transaction.TrackingStatus,
		TrackingUrlProvider:  transaction.TrackingURLProvider,
		Eta:                  timeops.ConvertToTimestampStringHour(transaction.Eta),
		LabelUrl:             transaction.LabelURL,
		CommercialInvoiceUrl: transaction.CommercialInvoiceURL,
	}

	s.Labels = append(s.Labels, label)

	return nil
}
