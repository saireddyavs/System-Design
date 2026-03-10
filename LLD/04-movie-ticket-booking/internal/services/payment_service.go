package services

import "movie-ticket-booking/internal/interfaces"

// MockPaymentProcessor implements PaymentProcessor for testing/demo
type MockPaymentProcessor struct{}

// NewMockPaymentProcessor creates a mock payment processor
func NewMockPaymentProcessor() interfaces.PaymentProcessor {
	return &MockPaymentProcessor{}
}

// ProcessPayment simulates payment processing
func (m *MockPaymentProcessor) ProcessPayment(amount float64, bookingID string) error {
	return nil
}

// ProcessRefund simulates refund processing
func (m *MockPaymentProcessor) ProcessRefund(amount float64, bookingID string) error {
	return nil
}
