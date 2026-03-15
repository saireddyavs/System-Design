package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"time"
)

var (
	ErrPaymentFailed = errors.New("payment processing failed")
)

// InMemoryPaymentProcessor implements PaymentProcessor for demo/testing
type InMemoryPaymentProcessor struct {
	paymentRepo interfaces.PaymentRepository
}

// NewInMemoryPaymentProcessor creates a new in-memory payment processor
func NewInMemoryPaymentProcessor(paymentRepo interfaces.PaymentRepository) *InMemoryPaymentProcessor {
	return &InMemoryPaymentProcessor{paymentRepo: paymentRepo}
}

// ProcessPayment processes a payment for a ride
func (p *InMemoryPaymentProcessor) ProcessPayment(rideID string, amount float64, method models.PaymentMethod) (*models.Payment, error) {
	b := make([]byte, 8)
	rand.Read(b)
	payment := &models.Payment{
		ID:            fmt.Sprintf("pay-%s-%s", rideID, hex.EncodeToString(b)),
		RideID:        rideID,
		Amount:        amount,
		Method:        method,
		Status:        models.PaymentStatusCompleted,
		TransactionID: fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	payment.CompletedAt = &now

	if err := p.paymentRepo.Create(payment); err != nil {
		return nil, err
	}
	return payment, nil
}
