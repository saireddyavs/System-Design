package interfaces

import (
	"context"
	"food-delivery-system/internal/models"
)

// PaymentProcessor defines the contract for payment processing (Interface Segregation)
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, payment *models.Payment) error
	RefundPayment(ctx context.Context, paymentID string) error
}
