package store

import (
	"crypto/md5"
	"encoding/hex"
)

const OrderStatusOpen = "OPEN"

const OrderStatusPaid = "PAID"

const OrderStatusShipped = "SHIPPED"

const OrderStatusDelivered = "DELIVERED"

const OrderStatusClosed = "CLOSED"

const OrderStatusOpenReturn = "OPEN_RETURN"

const OrderStatusRefunded = "OPEN_RETURN_REFUNDED"

const OrderStatusItemsReceived = "OPEN_RETURN_ITEMS_RECEIVED"

const OrderStatusReturned = "RETURNED"

// ShoppingCart represents a user's shopping cart.
type ShoppingCart struct {
	UserID        string               `json:"user_id"`
	Items         map[string]*CartItem `json:"items"` // item ID: CartItem
	TotalItems    int                  `json:"total_items"`
	Subtotal      float32              `json:"subtotal"` // sum of CartItems[i].ItemSubtotal
	CartWeightOzs float32              `json:"cart_weight_ozs"`
	CartWeightLbs float32              `json:"cart_weight_lbs"`
	CartWeightKgs float32              `json:"cart_weight_kgs"`
}

// CartItem represents a StoreItem added to user's cart for purchase.
// CartItems are not persisted outside of the ShoppingCart struct.
type CartItem struct {
	UserID         string  `json:"user_id"`
	ItemID         string  `json:"item_id"`
	SizeID         string  `json:"size_id"` // <itemID>-<size> (ex: '0001-xl')
	Name           string  `json:"name"`
	Size           string  `json:"size"`
	Quantity       int     `json:"quantity"`
	Price          float32 `json:"price"`
	ItemSubtotal   float32 `json:"item_subtotal"` // quantity * price
	UnitWeightOzs  float32 `json:"unit_weight_ozs"`
	TotalWeightOzs float32 `json:"total_weight_ozs"` // quantity * unit weight
	UnitWeightLbs  float32 `json:"unit_weight_lbs"`
	TotalWeightLbs float32 `json:"total_weight"` // quantity * unit weight
	ThumbnailID    string  `json:"thumbnail_id"`
}

// StoreItem represents an item available for purchase in the online store.
type StoreItem struct {
	ItemID         string         `json:"item_id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Category       string         `json:"category"`
	Subcategory    string         `json:"subcategory"`
	Price          float32        `json:"price"`
	UnitsSold      int            `json:"units_sold"`
	UnitsAvailable map[string]int `json:"units_available"` // size: units
	UnitWeightOzs  float32        `json:"unit_weight_ozs"`
	UnitWeightLbs  float32        `json:"unit_weight_lbs"`
	DateAdded      string         `json:"date_added"`
}

// StoreItemSummary contains summarized info of each StoreItem that is displayed when
// a user is browsing a selection of items.
type StoreItemSummary struct {
	ItemID      string  `json:"item_id"`
	Subcategory string  `json:"subcategory"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	ThumbnailID string  `json:"thumbnail_id"`
}

// Transaction represents a monetary transaction between the store and a user.
type Transaction struct {
	TransactionID     string  `json:"transaction_id"`
	OrderID           string  `json:"order_id"`
	UserID            string  `json:"user_id"`
	Timestamp         string  `json:"timestamp"`
	PaymentMethod     string  `json:"payment_method"`
	PaymentTxID       string  `json:"payment_tx_id"`
	SalesSubtotal     float32 `json:"sales_subtotal"`
	ShippingCost      float32 `json:"shipping_cost"`
	SalesTax          float32 `json:"sales_tax"`
	ChargesAndFees    float32 `json:"charges_and_fees"` // stripe processing, other fees
	TotalAmount       float32 `json:"total_amount"`
	PaymentStatus     string  `json:"payment_status"` // SUCCESS, FAILED, DISPUTED, REFUNDED
	PaymentMessage    string  `json:"payment_message"`
	CorrespondingTxID string  `json:"corresponding_tx_id"` // link to corresponding transaction for refunds
}

// SetHashID sets the t.TransctionID field with a MD5 hash generated from the t.Timestamp value
func (t *Transaction) SetHashID() {
	if t.Timestamp == "" {
		t.TransactionID = ""
		return
	}
	hash := md5.Sum([]byte(t.Timestamp))
	hashStr := hex.EncodeToString(hash[:])
	t.TransactionID = hashStr
	return
}

// Order represents a customer order for a store item.
type Order struct {
	OrderID          string      `json:"order_id"`
	TransactionID    string      `json:"transaction_id"`
	StripeChargeID   string      `json:"stripe_charge_id"`
	UserID           string      `json:"user_id"`
	Complete         bool        `json:"status"`    // denotes whether order is complete after creation at initial checkout page
	Expired          bool        `json:"expired"`   // denotes whether order is expired (checkout timeout / cart updated)
	TtlMs            int         `json:"ttl_ms"`    // time to live in ms
	InitTime         string      `json:"init_time"` // timestamp when order is created - format to/from time.Time obj
	Items            []*CartItem `json:"items"`
	TotalItems       int         `json:"total_items"` // sum of quantities of all items in cart
	SalesSubtotal    float32     `json:"sales_subtotal"`
	ShippingCost     float32     `json:"shipping_cost"`
	SalesTax         float32     `json:"sales_tax"`
	ChargesAndFees   float32     `json:"charges_and_fees"` // stripe processing, other fees
	OrderTotal       float32     `json:"order_total"`
	OrderDate        string      `json:"order_date"`
	TxTimestamp      string      `json:"transaction_timestamp"`
	PaymentStatus    string      `json:"payment_status"`
	Paid             bool        `json:"paid"`
	BillingAddress   Address     `json:"billing_address"`  // Address, City, State, ZIP
	ShippingAddress  Address     `json:"shipping_address"` // Address, City, State, ZIP
	OrderWeightOzs   float32     `json:"order_weight_ozs"`
	OrderWeightLbs   float32     `json:"order_weight_lbs"`
	OrderWeightKgs   float32     `json:"order_weight_kg"`
	Shipped          bool        `json:"shipped"`
	ShippingCarriers []string    `json:"shipping_carrier"` // []string array types for large, multi-part, and international orders
	ShippingMethods  []string    `json:"shipping_method"`
	ShipDates        []string    `json:"ship_date"`
	TrackingNumbers  []string    `json:"tracking_number"`
	Delivered        bool        `json:"delivered"`
	OrderStatus      string      `json:"order_status"`
}

// Receipt represents a receipt sent to customers after placing orders.
type Receipt struct {
	UserID          string      `json:"user_id"`
	OrderID         string      `json:"order_id"`
	TransactionID   string      `json:"transaction_id"`
	UserEmail       string      `json:"user_email"`
	OrderSummary    []*CartItem `json:"order_summary"`
	SalesSubtotal   float32     `json:"sales_subtotal"`
	ShippingCost    float32     `json:"shipping_cost"`
	SalesTax        float32     `json:"sales_tax"`
	ChargesAndFees  float32     `json:"charges_and_fees"` // stripe processing, other fees
	OrderTotal      float32     `json:"order_total"`
	BillingAddress  Address     `json:"billing_address"`  // Address, City, State, ZIP
	ShippingAddress Address     `json:"shipping_address"` // Address, City, State, ZIP
}

// New sets the values of a Receipt object with the given
// data from the *Customer and *Order objects.
func (r *Receipt) New(cust *Customer, order *Order) {
	r.UserID = cust.UserID
	r.OrderID = order.OrderID
	r.TransactionID = order.TransactionID
	r.UserEmail = cust.Email
	r.OrderSummary = order.Items
	r.SalesSubtotal = order.SalesSubtotal
	r.ShippingCost = order.ShippingCost
	r.SalesTax = order.SalesTax
	r.ChargesAndFees = order.ChargesAndFees
	r.OrderTotal = order.OrderTotal
	r.BillingAddress = order.BillingAddress
	r.ShippingAddress = order.ShippingAddress
}

// Return represents a customer return request
type Return struct {
	UserID        string         `json:"user_id"`
	UserEmail     string         `json:"user_email"`
	ReturnID      string         `json:"return_id"`
	OrderID       string         `json:"order_id"`
	OriginalTxID  string         `json:"original_tx_id"`
	ReturnTxID    string         `json:"return_tx_id"`
	ReturnItems   map[string]int `json:"return_items"` // return item IDs: quantity
	Open          bool           `json:"open"`
	Refunded      bool           `json:"refunded"`
	ItemsReceived bool           `json:"items_received"`
	Complete      bool           `json:"complete"`
	ReturnStatus  string         `json:"return_status"`
}

// Customer represents a user of the service.
type Customer struct {
	UserID          string  `json:"user_id"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	BillingAddress  Address `json:"billing_address"`
	ShippingAddress Address `json:"shipping_address"`
	City            string  `json:"city"`
	Country         string  `json:"country"`
	Purchases       int     `json:"purchases"`     // total number of purchases
	Returns         int     `json:"returns"`       // total number of returns
	Disputes        int     `json:"disputes"`      // total number of disputes
	TotalSpent      float32 `json:"total_spent"`   // total USD spent
	Orders          int     `json:"orders"`        // total number of orders created
	OpenOrder       bool    `json:"open_order"`    // denotes if customer has order in progress
	OpenOrderID     string  `json:"open_order_id"` // ID of open order
	JoinDate        string  `json:"join_date"`
}

// Address represents a mailling or billing address.
type Address struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Company      string `json:"company"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	Zip          string `json:"zip"`
	PhoneNumber  string `json:"phone_number"`
}

// TxStatusFail contains the status value for failed Transactions.
const TxStatusFail = "FAILED"

// TxStatusComplete contains the status value for successful Transactions.
const TxStatusComplete = "SUCCESS"

// TxStatusDispute containst the status value for disputed Transactions
const TxStatusDispute = "DISPUTED"

// TxStatusRefund contains the status value for refunded Transactions
const TxStatusRefund = "REFUNDED"

// temporary cache - pull from DynamoDB in prod
/* var shoppingCarts = make(map[string]ShoppingCart)

func AddToCart(item CartItem, cart *ShoppingCart) {
	cart.Items = append(cart.Items, item)
	cart.TotalItems += item.Quantity
	cart.Subtotal += item.ItemSubtotal
} */
