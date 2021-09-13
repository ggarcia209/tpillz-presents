package dbops

import (
	"fmt"
	"log"
	"sync"

	"github.com/go-aws/go-dynamo/dynamo"
	"github.com/tpillz-presents/service/store-api/store"
)

// CustomersTable contains the name of the Users Table.
const CustomersTable = "tpillz-customers-dev"

// CustomersPK contains the primary key name of the Users Table.
const CustomersPK = "email"

// StoreItemsTable contains the name of the StoreItems Table.
const StoreItemsTable = "tpillz-store-items-dev"

// StoreItemPK contains the primary key name of the StoreItems Table.
const StoreItemPK = "sub_category"

// StoreItemSK contains the sort key name of the StoreItems Table.
const StoreItemSK = "item_id"

// StoreItemsSummaryTable contains the name of the StoreItemsSumary Table.
const StoreItemsSummaryTable = "tpillz-store-items-summary-dev"

// StoreItemSummaryPK contains the primary key name of the StoreItemsSumary Table.
const StoreItemSummaryPK = "sub_category"

// StoreItemSummarySK contains the sort key name of the StoreItemsSumary Table.
const StoreItemSummarySK = "item_id"

// ShoppingCartsTable contains the name of the ShoppingCarts Table.
const ShoppingCartsTable = "tpillz-shopping-carts-dev"

// ShoppingCartsPK contains the primary key name of the ShoppingCarts Table.
const ShoppingCartsPK = "user_id"

// ShippingMethodsTable contains the name of the ShippingMethods table.
const ShippingMethodsTable = "tpillz-shipping-methods"

// ShippingMethodsPK contains the primary key name of the ShippingMethods table.
const ShippingMethodsPK = "method_name"

// OrdersTable contains the name of the Orders table.
const OrdersTable = "tpillz-orders-dev"

// OrdersPK contains the primary key name of the Orders table.
const OrdersPK = "user_id"

// OrdersSK contains the sort key name of the Orders table.
const OrdersSK = "order_id"

// TransactionsTable contains the name of the Transactions Table.
const TransactionsTable = "tpillz-transactions-dev"

// TransactionsPK contains the primary key name of the Transactions Table.
const TransactionsPK = "user_id"

// TransactionsSK contains the sort key name of the Transactions Table.
const TransactionsSK = "transaction_id"

// ParcelsTable contains the table name of the parcels table - contains parcel data used for shipping.
const ParcelsTable = "tpillz-parcels-dev"

const ParcelsPK = "carrier"

const ParcelsSK = "parcel_id"

// ShipmentsTable contains the table name of the Shipments table
// containing information about order shipments used for order fullfillment.
const ShipmentsTable = "tpillz-shipments-dev"

const ShipmentsPK = "user_id"

const ShipmentsSK = "order_id"

// ErrConditionCheckFail contains the error code values for failed conditional writes.
const ErrConditionalCheck = "ERR_CONDITIONAL_CHECK"

// Table contains the necessary information to access the service's DynamoDB tables.
// Primary & Sort key types are hardcoded as string format.
type Table struct {
	Name       string
	PrimaryKey string
	SortKey    string
}

// Construct sets the Table object's fields with the given values.
func (t *Table) Construct(name, primaryKey, sortKey string) {
	t.Name = name
	t.PrimaryKey = primaryKey
	t.SortKey = sortKey
}

// NewTable returns a new Table object per the given arguments.
func NewTable(name, primaryKey, sortKey string) Table {
	t := Table{
		Name:       name,
		PrimaryKey: primaryKey,
		SortKey:    sortKey,
	}
	return t
}

// InitDB initializes a new DynamoDB session and creates a dynamo.DbInfo object with
// the defined Table objects to be used by the program.
func InitDB(tables []Table) *dynamo.DbInfo {
	svc := dynamo.InitSesh()
	db := dynamo.InitDbInfo()
	db.SetSvc(svc)
	for _, table := range tables {
		t := dynamo.CreateNewTableObj(table.Name, table.PrimaryKey, "string", table.SortKey, "string")
		db.AddTable(t)
	}

	return db
}

// GetStoreItem retreives a StoreItem object from the StoreItemsTable.
func GetStoreItem(DB *dynamo.DbInfo, subcategory, itemID string) (*store.StoreItem, error) {
	q := dynamo.CreateNewQueryObj(subcategory, itemID)
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[StoreItemsTable], &store.StoreItem{}, expr)
	if err != nil {
		log.Printf("GetStoreItem failed: %v", err)
		return &store.StoreItem{}, err
	}
	return item.(*store.StoreItem), nil
}

// PutStoreItem puts a new StoreItem object to the StoreItemsTable.
func PutStoreItem(DB *dynamo.DbInfo, item *store.StoreItem) error {
	err := dynamo.CreateItem(DB.Svc, item, DB.Tables[StoreItemsTable])
	if err != nil {
		log.Printf("PutStoreItem failed: %v", err)
	}
	return nil
}

// GetStoreItemSummary retreives a StoreItemSummary object from the StoreItemsSummaryTable.
func GetStoreItemSummmary(DB *dynamo.DbInfo, subcategory, itemID string) (*store.StoreItemSummary, error) {
	q := dynamo.CreateNewQueryObj(subcategory, itemID)
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[StoreItemsSummaryTable], &store.StoreItemSummary{}, expr)
	if err != nil {
		log.Printf("GetStoreItemSummmary failed: %v", err)
		return &store.StoreItemSummary{}, err
	}
	return item.(*store.StoreItemSummary), nil
}

// PutStoreItemSummary puts a new StoreItemSummary object to the StoreItemsSummaryTable.
func PutStoreItemSummmary(DB *dynamo.DbInfo, item *store.StoreItemSummary) error {
	err := dynamo.CreateItem(DB.Svc, item, DB.Tables[StoreItemsSummaryTable])
	if err != nil {
		log.Printf("PutStoreItemSummmary failed: %v", err)
	}
	return nil
}

// GetShopping cart retreives a ShoppingCart object from the ShoppingCartsTable (primary key only).
func GetShoppingCart(DB *dynamo.DbInfo, userID string) (*store.ShoppingCart, error) {
	q := dynamo.CreateNewQueryObj(userID, "")
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[ShoppingCartsTable], &store.ShoppingCart{}, expr)
	if err != nil {
		log.Printf("GetShoppingCart failed: %v", err)
		return &store.ShoppingCart{}, err
	}
	if len(item.(*store.ShoppingCart).Items) == 0 {
		item.(*store.ShoppingCart).Items = make(map[string]*store.CartItem)
	}
	return item.(*store.ShoppingCart), nil
}

// PutShoppingCart puts a new ShoppingCart object to the ShoppingCartsTable.
func PutShoppingCart(DB *dynamo.DbInfo, cart *store.ShoppingCart) error {
	err := dynamo.CreateItem(DB.Svc, cart, DB.Tables[ShoppingCartsTable])
	if err != nil {
		log.Printf("PutShoppingCart failed: %v", err)
	}
	return nil
}

// PutParcel adds a new store.Parcel object to the Parcels table.
func PutParcel(DB *dynamo.DbInfo, parcel []*store.Parcel) error {
	err := dynamo.CreateItem(DB.Svc, parcel, DB.Tables[ParcelsTable])
	if err != nil {
		log.Printf("PutShoppingCart failed: %v", err)
	}
	return nil
}

// GetParcels scans the Parcel table for all options matching the given carrier
func GetParcels(DB *dynamo.DbInfo, carrier string) ([]*store.Parcel, error) {
	parcels := []*store.Parcel{}

	eb := dynamo.NewExprBuilder()
	eb.SetFilter("carrier", carrier)
	expr, err := eb.BuildExpression()
	if err != nil {
		log.Printf("GetParcels failed: %v", err)
		return parcels, err
	}

	items, err := dynamo.ScanItems(DB.Svc, DB.Tables[ParcelsTable], &store.Parcel{}, expr)
	if err != nil {
		log.Printf("GetParcels failed: %v", err)
		return parcels, err
	}

	for _, item := range items {
		parcels = append(parcels, item.(*store.Parcel))
	}

	return parcels, nil
}

// PutShipment puts a new Shipment object to the ShipmentsTable.
func PutShipment(DB *dynamo.DbInfo, shipment *store.Shipment) error {
	err := dynamo.CreateItem(DB.Svc, shipment, DB.Tables[ShipmentsTable])
	if err != nil {
		log.Printf("PutShipment failed: %v", err)
	}
	return nil
}

func GetShippingMethod(DB *dynamo.DbInfo, methodName string) (*store.ShippingMethod, error) {
	q := dynamo.CreateNewQueryObj(methodName, "")
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[ShippingMethodsTable], &store.ShippingMethod{}, expr)
	if err != nil {
		log.Printf("GetShippingMethod failed: %v", err)
		return &store.ShippingMethod{}, err
	}
	return item.(*store.ShippingMethod), nil
}

// PutShoppingCart puts a new ShoppingCart object to the ShoppingCartsTable.
func PutShippingMethod(DB *dynamo.DbInfo, method *store.ShippingMethod) error {
	err := dynamo.CreateItem(DB.Svc, method, DB.Tables[ShippingMethodsTable])
	if err != nil {
		log.Printf("PutShippingMethod failed: %v", err)
	}
	return nil
}

// GetTransaction retreives a Transaction object from the TransactionsTable.
func GetTransaction(DB *dynamo.DbInfo, userID, txID string) (*store.Transaction, error) {
	q := dynamo.CreateNewQueryObj(userID, txID)
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[TransactionsTable], &store.Transaction{}, expr)
	if err != nil {
		log.Printf("GetTransaction failed: %v", err)
		return &store.Transaction{}, err
	}
	return item.(*store.Transaction), nil
}

// PutTransaction puts a new Transaction object to the TransactionsTable.
func PutTransaction(DB *dynamo.DbInfo, tx *store.Transaction) error {
	err := dynamo.CreateItem(DB.Svc, tx, DB.Tables[TransactionsTable])
	if err != nil {
		log.Printf("PutTransaction failed: %v", err)
	}
	return nil
}

// GetCustomer retreives a Customer object from the CustomersTable (primary key only).
func GetCustomer(DB *dynamo.DbInfo, email string) (*store.Customer, error) {
	q := dynamo.CreateNewQueryObj(email, "")
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[CustomersTable], &store.Customer{}, expr)
	if err != nil {
		log.Printf("GetCustomer failed: %v", err)
		return &store.Customer{}, err
	}
	return item.(*store.Customer), nil
}

// PutCustomer puts a new Customer object to the CustomersTable.
func PutCustomer(DB *dynamo.DbInfo, user *store.Customer) error {
	err := dynamo.CreateItem(DB.Svc, user, DB.Tables[CustomersTable])
	if err != nil {
		log.Printf("PutCustomer failed: %v", err)
	}
	return nil
}

// GetOrder retreives an Order object from the Orders table.
func GetOrder(DB *dynamo.DbInfo, userID, orderID string) (*store.Order, error) {
	q := dynamo.CreateNewQueryObj(userID, orderID)
	expr := dynamo.NewExpression()
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[OrdersTable], &store.Order{}, expr)
	if err != nil {
		log.Printf("GetOrder failed: %v", err)
		return &store.Order{}, err
	}
	return item.(*store.Order), nil
}

// GetOrderItems retreives an Order object's Items from the Orders table.
func GetOrderItems(DB *dynamo.DbInfo, userID, orderID string) (*store.Order, error) {
	q := dynamo.CreateNewQueryObj(userID, orderID)
	eb := dynamo.NewExprBuilder()
	eb.SetProjection([]string{"items"})
	expr, err := eb.BuildExpression()
	if err != nil {
		log.Printf("GetOrderItems failed: %v", err)
		return &store.Order{}, err
	}
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[OrdersTable], &store.Order{}, expr)
	if err != nil {
		log.Printf("GetOrderItems failed: %v", err)
		return &store.Order{}, err
	}
	return item.(*store.Order), nil
}

// PutOrder puts a new Order object to the Orders table.
func PutOrder(DB *dynamo.DbInfo, user *store.Order) error {
	err := dynamo.CreateItem(DB.Svc, user, DB.Tables[OrdersTable])
	if err != nil {
		log.Printf("PutOrder failed: %v", err)
	}
	return nil
}

func UpdateOrderAddress(DB *dynamo.DbInfo, userID, orderID string, addr store.Address, shipping bool) error {
	field := "shipping_address"
	if !shipping {
		field = "billing_address"
	}

	// create and set update query
	q := dynamo.CreateNewQueryObj(userID, orderID)
	q.UpdateCurrent(field, addr)

	// build expression
	update := dynamo.NewUpdateExpr()
	update.Set(field, addr)

	eb := dynamo.NewExprBuilder()
	eb.SetUpdate(update)
	expression, err := eb.BuildExpression()
	if err != nil {
		log.Printf("UpdateOrderAddress failed: %v", err)
		return err
	}

	// update DB object
	err = dynamo.UpdateItem(DB.Svc, q, DB.Tables[OrdersTable], expression)
	if err != nil {
		log.Printf("UpdateOrderAddress failed: %v", err)
		return err
	}
	return nil

}

func UpdateOrderPaymentStatus(DB *dynamo.DbInfo, customerID, orderID, status string) error {
	q := dynamo.CreateNewQueryObj(customerID, orderID)
	expr := dynamo.NewExpression()
	q.UpdateCurrent("payment_status", status)
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[OrdersTable], expr)
	if err != nil {
		log.Printf("UpdateOrderPaymentStatus failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentStatus(DB *dynamo.DbInfo, customerID, txID, status string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_status", status)
	expr := dynamo.NewExpression()
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable], expr)
	if err != nil {
		log.Printf("UpdateOrderPaymentStatus failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentMethod(DB *dynamo.DbInfo, customerID, txID, method string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_method", method)
	expr := dynamo.NewExpression()
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable], expr)
	if err != nil {
		log.Printf("UpdateOrderPaymentMethod failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentID(DB *dynamo.DbInfo, customerID, txID, paymentID string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_tx_id", paymentID)
	expr := dynamo.NewExpression()
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable], expr)
	if err != nil {
		log.Printf("UpdateOrderPaymentID failed: %v", err)
		return err
	}
	return nil
}

// VerifyOrderStock verifies that all items in an order are still in stock at the time of payment.
func VerifyOrderStock(DB *dynamo.DbInfo, items []*store.CartItem) (bool, []string, error) {
	bc := make(chan map[string]bool)
	ec := make(chan error)
	var wg sync.WaitGroup

	ok := true
	out := []string{} // list of out of stock items

	for _, item := range items {
		wg.Add(1)
		go checkInventory(DB, item, bc, ec, &wg)
	}

	br := 0
	er := 0
	for {
		select {
		case check := <-bc:
			br++
			for k, v := range check {
				if !v {
					out = append(out, k)
					ok = false
				}
			}
		case err := <-ec:
			er++
			if err != nil {
				log.Printf("VerifyOrderStock failed: %v", err)
				return false, []string{}, err
			}
		}
		if er == len(items) && br == len(items) {
			break
		}
	}

	wg.Wait()
	close(bc)
	close(ec)

	return ok, out, nil
}

// checkInventory runs as a goroutine to check the availability of a shopping cart's items concurrently
func checkInventory(DB *dynamo.DbInfo, item *store.CartItem, bc chan map[string]bool, ec chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	field := "units_available"
	keyName := fmt.Sprintf("%s.%s", field, item.Size)
	q := dynamo.CreateNewQueryObj(item.Subcategory, item.ItemID)

	eb := dynamo.NewExprBuilder()
	eb.SetProjection([]string{keyName})
	expr, err := eb.BuildExpression()

	check, err := dynamo.GetItem(DB.Svc, q, DB.Tables[StoreItemsTable], &store.StoreItem{}, expr)
	if err != nil {
		log.Printf("checkInventory failed: %v", err)
		bc <- map[string]bool{item.ItemID: false}
		ec <- err
		return
	}
	if check.(*store.StoreItem).UnitsAvailable[item.Size] < item.Quantity {
		bc <- map[string]bool{item.ItemID: false}
		ec <- nil
		return
	}
	bc <- map[string]bool{item.ItemID: true}
	ec <- nil
	return
}

// UpdateInventoryCount updates a Store Item's inventory count by size and decrements the value for
// the given sizeKey by the count integer. The update succeeds on the condition that the quantity
// of the given size is greater than or equal to the count variable. Returns ItemID and ConditionalCheck error if item
// is out of stock.
func UpdateInventoryCount(DB *dynamo.DbInfo, subcat, itemID, sizeKey string, count int) (string, error) {
	field := "units_available"
	keyName := fmt.Sprintf("%s.%s", field, sizeKey)

	// create and set update query
	q := dynamo.CreateNewQueryObj(subcat, itemID)
	q.UpdateCurrent(field, count)

	// build expression
	cond := dynamo.NewCondition()
	cond.GreaterThanEqual(keyName, count)

	update := dynamo.NewUpdateExpr()
	update.SetMinus(keyName, keyName, count, true)

	eb := dynamo.NewExprBuilder()
	eb.SetCondition(cond)
	eb.SetUpdate(update)
	expression, err := eb.BuildExpression()
	if err != nil {
		log.Printf("UpdateInventoryCount failed: %v", err)
		return "", err
	}

	err = dynamo.UpdateItem(DB.Svc, q, DB.Tables[StoreItemsTable], expression)
	if err != nil {
		if err.Error() == dynamo.ErrConditionalCheck {
			return itemID, fmt.Errorf(ErrConditionalCheck)
		}
		log.Printf("UpdateInventoryCount failed: %v", err)
		return "", err
	}
	return "", nil
}
