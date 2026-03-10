package strategies

import (
	"airline-reservation-system/internal/interfaces"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrPaymentFailed = errors.New("payment failed")
)

// MockPaymentProcessor implements PaymentProcessor for testing (no real payments)
type MockPaymentProcessor struct {
	transactions map[string]float64
	mu           sync.RWMutex
}

// NewMockPaymentProcessor creates a mock payment processor
func NewMockPaymentProcessor() interfaces.PaymentProcessor {
	return &MockPaymentProcessor{
		transactions: make(map[string]float64),
	}
}

func (m *MockPaymentProcessor) ProcessPayment(amount float64, currency string, reference string) (string, error) {
	if amount <= 0 {
		return "", ErrPaymentFailed
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	txnID := fmt.Sprintf("TXN-%s-%d", reference, len(m.transactions))
	m.transactions[txnID] = amount
	return txnID, nil
}

func (m *MockPaymentProcessor) ProcessRefund(transactionID string, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.transactions[transactionID]; !exists {
		return errors.New("transaction not found")
	}
	delete(m.transactions, transactionID)
	return nil
}
