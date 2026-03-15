package models

import "time"

// PaymentMethod represents the type of payment
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "CreditCard"
	PaymentMethodDebitCard  PaymentMethod = "DebitCard"
	PaymentMethodUPI        PaymentMethod = "UPI"
	PaymentMethodWallet     PaymentMethod = "Wallet"
)

// PaymentStatus represents the payment state
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "Pending"
	PaymentStatusCompleted PaymentStatus = "Completed"
)

// Payment represents a payment transaction
type Payment struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	Amount        float64       `json:"amount"`
	Method        PaymentMethod `json:"method"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
