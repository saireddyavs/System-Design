package services

import (
	"context"
	"fmt"
	"food-delivery-system/internal/models"
	"sync"
	"time"
)

// InMemoryPaymentProcessor simulates payment processing
type InMemoryPaymentProcessor struct {
	payments map[string]*models.Payment
	mu       sync.RWMutex
}

// NewInMemoryPaymentProcessor creates a new in-memory payment processor
func NewInMemoryPaymentProcessor() *InMemoryPaymentProcessor {
	return &InMemoryPaymentProcessor{
		payments: make(map[string]*models.Payment),
	}
}

// ProcessPayment simulates successful payment processing
func (p *InMemoryPaymentProcessor) ProcessPayment(ctx context.Context, payment *models.Payment) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	payment.TransactionID = fmt.Sprintf("TXN-%d", time.Now().UnixNano())
	payment.Status = models.PaymentStatusCompleted
	now := time.Now()
	payment.CompletedAt = &now
	p.payments[payment.ID] = payment
	return nil
}

// RefundPayment simulates refund
func (p *InMemoryPaymentProcessor) RefundPayment(ctx context.Context, paymentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	payment, exists := p.payments[paymentID]
	if !exists {
		return fmt.Errorf("payment not found: %s", paymentID)
	}
	payment.Status = models.PaymentStatusRefunded
	return nil
}
