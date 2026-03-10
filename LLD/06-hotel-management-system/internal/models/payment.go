package models

import (
	"sync"
	"time"
)

// PaymentStatus represents payment state
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "Pending"
	PaymentStatusCompleted PaymentStatus = "Completed"
	PaymentStatusFailed    PaymentStatus = "Failed"
	PaymentStatusRefunded  PaymentStatus = "Refunded"
)

// PaymentMethod represents how payment was made
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "Cash"
	PaymentMethodCard     PaymentMethod = "Card"
	PaymentMethodUPI      PaymentMethod = "UPI"
	PaymentMethodBankTransfer PaymentMethod = "BankTransfer"
)

// Payment represents a payment transaction
type Payment struct {
	ID            string
	BookingID     string
	Amount        float64
	Method        PaymentMethod
	Status        PaymentStatus
	TransactionID string
	PaidAt        *time.Time
	RefundAmount  float64
	mu            sync.RWMutex
}

// NewPayment creates a new Payment instance
func NewPayment(id, bookingID string, amount float64, method PaymentMethod) *Payment {
	return &Payment{
		ID:        id,
		BookingID: bookingID,
		Amount:    amount,
		Method:    method,
		Status:    PaymentStatusPending,
	}
}

// GetStatus returns payment status (thread-safe)
func (p *Payment) GetStatus() PaymentStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status
}

// SetStatus updates payment status (thread-safe)
func (p *Payment) SetStatus(status PaymentStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = status
	if status == PaymentStatusCompleted && p.PaidAt == nil {
		now := time.Now()
		p.PaidAt = &now
	}
}

// GetAmount returns payment amount (thread-safe)
func (p *Payment) GetAmount() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Amount
}
