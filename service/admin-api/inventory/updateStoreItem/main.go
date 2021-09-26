package main

/* updateStoreItem updates a specified field of a StoreItem object. */

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/tpillz-presents/service/util/dbops"
	"github.com/tpillz-presents/service/util/httpops"
)

type updateReq struct {
	Subcategory string      `json:"sub_category"`
	ItemID      string      `json:"item_id"`
	FieldName   string      `json:"field_name"`
	Value       interface{} `json:"value"`
	ValueType   string      `json:"value_type"`
}

const route = "/admin/inventory/update_item" // POST
const failMsg = "Request failed!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{
		Name:       dbops.StoreItemsTable(),
		PrimaryKey: dbops.StoreItemPK,
		SortKey:    dbops.StoreItemSK,
	},
	dbops.Table{
		Name:       dbops.StoreItemsSummaryTable(),
		PrimaryKey: dbops.StoreItemSummaryPK,
		SortKey:    dbops.StoreItemSummarySK,
	},
}

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	DB := dbops.InitDB(tables)

	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := updateReq{}
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

	// get put new store item to DB
	err = dbops.UpdateStoreItem(DB, data.Subcategory, data.ItemID, data.FieldName, data.Value)
	if err != nil {
		log.Printf("RootHandler failed: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// update ItemSummary object if needed
	updateSummary := map[string]bool{
		"name":        true,
		"subcategory": true,
		"price":       true,
	}
	if updateSummary[data.FieldName] {
		err = dbops.UpdateStoreItemSummary(DB, data.Subcategory, data.ItemID, data.FieldName, data.Value)
		if err != nil {
			log.Printf("RootHandler failed: %v", err)
			httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
			return
		}
	}

	// return order to admin
	httpops.ErrResponse(w, "Success! Item added!", data.Value, http.StatusOK)
	return
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
