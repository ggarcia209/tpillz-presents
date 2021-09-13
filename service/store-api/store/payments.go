package store

// PaymentStatusSuccess contains the status code for successful transactions.
const PaymentStatusSuccess = "PAYMENT_SUCCESS"

// PaymentStatusFail contains the status code for failed transactions.
const PaymentStatusFail = "PAYMENT_FAIL"

// PaymentStatusInProgress contains the status code for in progress transactions.
const PaymentStatusInProgress = "IN_PROGRESS"

// RefundStatusSuccess contains the status code for successful refunds.
const RefundStatusSuccess = "REFUND_SUCCESS"

// RefundStatusFail contains the status code for failed refunds.
const RefundStatusFail = "REFUND_FAIL"

const PaymentDisputed = "DISPUTED"

// ValidPaymentStatus contains the valid values for Stripe transaction statuses.
var ValidPaymentStatus = map[string]bool{
	PaymentStatusFail:       true,
	PaymentStatusSuccess:    true,
	PaymentStatusInProgress: true,
	RefundStatusFail:        true,
	RefundStatusSuccess:     true,
}

// PaymentStatus contains message info to send to the PaymentStatus fifo queue.
// Objects staged in the Staging fifo queue are processed on receipt of this message
type PaymentStatus struct {
	CustomerEmail string `json:"customer_email"`
	CustomerID    string `json:"customer_id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	PaymentMethod string `json:"payment_method"` // 3rd party payment platform name (ex: Stripe, Apple Pay)
	PaymentTxID   string `json:"payment_tx_id"`  // 3rd party payment transaction ID returned by API
	TxStatus      string `json:"tx_status"`
	TxMessage     string `json:"tx_message"`
}
