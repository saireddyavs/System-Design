package models

import "time"

// PaymentStatus represents payment state
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod represents how payment is made
type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodUPI    PaymentMethod = "upi"
	PaymentMethodCash   PaymentMethod = "cash"
	PaymentMethodWallet PaymentMethod = "wallet"
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
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
}

// NewPayment creates a new payment
func NewPayment(id, orderID string, amount float64, method PaymentMethod) *Payment {
	return &Payment{
		ID:        id,
		OrderID:   orderID,
		Amount:    amount,
		Method:    method,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
	}
}
