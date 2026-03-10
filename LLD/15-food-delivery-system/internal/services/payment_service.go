package services

import (
	"context"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
)

// PaymentService wraps the payment processor
type PaymentService struct {
	processor interfaces.PaymentProcessor
}

// NewPaymentService creates a new payment service
func NewPaymentService(processor interfaces.PaymentProcessor) *PaymentService {
	return &PaymentService{processor: processor}
}

// ProcessPayment processes a payment
func (s *PaymentService) ProcessPayment(ctx context.Context, payment *models.Payment) error {
	return s.processor.ProcessPayment(ctx, payment)
}
