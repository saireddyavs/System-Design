package services

import (
	"context"
	"errors"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

var ErrUnsupportedPaymentMethod = errors.New("unsupported payment method")

// PaymentService processes payments using Strategy pattern
type PaymentService struct {
	mu         sync.RWMutex
	processors map[models.PaymentMethod]interfaces.PaymentProcessor
}

// NewPaymentService creates a payment service with processors
func NewPaymentService(processors []interfaces.PaymentProcessor) *PaymentService {
	procMap := make(map[models.PaymentMethod]interfaces.PaymentProcessor)
	for _, p := range processors {
		procMap[p.GetMethod()] = p
	}
	return &PaymentService{
		processors: procMap,
	}
}

// ProcessPayment routes to the appropriate payment processor
func (s *PaymentService) ProcessPayment(ctx context.Context, payment *models.Payment) error {
	s.mu.RLock()
	processor, ok := s.processors[payment.Method]
	s.mu.RUnlock()

	if !ok {
		return ErrUnsupportedPaymentMethod
	}
	return processor.Process(ctx, payment)
}

