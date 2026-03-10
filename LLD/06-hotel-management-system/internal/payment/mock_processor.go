package payment

import (
	"fmt"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
	"time"
)

// MockPaymentProcessor simulates payment processing for testing
type MockPaymentProcessor struct {
	transactions map[string]bool
	mu           sync.RWMutex
}

// NewMockPaymentProcessor creates a mock processor
func NewMockPaymentProcessor() interfaces.PaymentProcessor {
	return &MockPaymentProcessor{
		transactions: make(map[string]bool),
	}
}

func (m *MockPaymentProcessor) ProcessPayment(payment *models.Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	payment.SetStatus(models.PaymentStatusCompleted)
	payment.TransactionID = fmt.Sprintf("TXN-%s", payment.ID)
	if payment.PaidAt == nil {
		t := time.Now()
		payment.PaidAt = &t
	}
	m.transactions[payment.ID] = true
	return nil
}

func (m *MockPaymentProcessor) ProcessRefund(payment *models.Payment, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	payment.RefundAmount = amount
	payment.SetStatus(models.PaymentStatusRefunded)
	return nil
}
