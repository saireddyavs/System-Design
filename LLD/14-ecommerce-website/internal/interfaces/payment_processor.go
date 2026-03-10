package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// PaymentProcessor defines the interface for payment processing (Strategy pattern)
type PaymentProcessor interface {
	Process(ctx context.Context, payment *models.Payment) error
	GetMethod() models.PaymentMethod
}
