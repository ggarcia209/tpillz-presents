package main

/* updateShipment updates a store.Shipment object in the DynamoDB Shipments table. */

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tpillz-presents/service/store-api/store"
	"github.com/tpillz-presents/service/util/dbops"
)

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{ // customers table
		Name:       dbops.ShipmentsTable,
		PrimaryKey: dbops.ShipmentsPK,
		SortKey:    dbops.ShipmentsSK,
	},
}

func handler(ctx context.Context, snsEvent events.SNSEvent) {
	db := dbops.InitDB(tables)
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		// fmt.Printf("[%s %s] Message = %s \n", record.EventSource, snsRecord.Timestamp, snsRecord.Message)
		msg := snsRecord.Message
		ship := &store.Shipment{}

		// unmarshall json string
		err := json.Unmarshal([]byte(msg), ship)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

		// update shipment in db
		err = dbops.PutShipment(db, ship)
		if err != nil {
			// handle err
			log.Printf("handler failed: %v", err)
			return
		}

	}
	return
}

func main() {
	lambda.Start(handler)
}
