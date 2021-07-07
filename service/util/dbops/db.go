package dbops

import (
	"log"

	"github.com/go-dynamo/dynamo"
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[StoreItemsTable], &store.StoreItem{})
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[StoreItemsSummaryTable], &store.StoreItemSummary{})
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[ShoppingCartsTable], &store.ShoppingCart{})
	if err != nil {
		log.Printf("GetShoppingCart failed: %v", err)
		return &store.ShoppingCart{}, err
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

func GetShippingMethod(DB *dynamo.DbInfo, methodName string) (*store.ShippingMethod, error) {
	q := dynamo.CreateNewQueryObj(methodName, "")
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[ShippingMethodsTable], &store.ShippingMethod{})
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[TransactionsTable], &store.Transaction{})
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[CustomersTable], &store.Customer{})
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
	item, err := dynamo.GetItem(DB.Svc, q, DB.Tables[OrdersTable], &store.Order{})
	if err != nil {
		log.Printf("GetOrder failed: %v", err)
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

func UpdateOrderPaymentStatus(DB *dynamo.DbInfo, customerID, orderID, status string) error {
	q := dynamo.CreateNewQueryObj(customerID, orderID)
	q.UpdateCurrent("payment_status", status)
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[OrdersTable])
	if err != nil {
		log.Printf("UpdateOrderPaymentStatus failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentStatus(DB *dynamo.DbInfo, customerID, txID, status string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_status", status)
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable])
	if err != nil {
		log.Printf("UpdateOrderPaymentStatus failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentMethod(DB *dynamo.DbInfo, customerID, txID, method string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_method", method)
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable])
	if err != nil {
		log.Printf("UpdateOrderPaymentMethod failed: %v", err)
		return err
	}
	return nil
}

func UpdateTxPaymentID(DB *dynamo.DbInfo, customerID, txID, paymentID string) error {
	q := dynamo.CreateNewQueryObj(customerID, txID)
	q.UpdateCurrent("payment_tx_id", paymentID)
	err := dynamo.UpdateItem(DB.Svc, q, DB.Tables[TransactionsTable])
	if err != nil {
		log.Printf("UpdateOrderPaymentID failed: %v", err)
		return err
	}
	return nil
}
