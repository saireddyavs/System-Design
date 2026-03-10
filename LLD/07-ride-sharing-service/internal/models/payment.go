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
	PaymentMethodWallet PaymentMethod = "wallet"
	PaymentMethodCash   PaymentMethod = "cash"
)

// Payment represents a ride payment
type Payment struct {
	ID            string
	RideID        string
	Amount        float64
	Method        PaymentMethod
	Status        PaymentStatus
	TransactionID string
	CreatedAt     time.Time
	CompletedAt   *time.Time
}
