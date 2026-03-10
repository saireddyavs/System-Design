package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// PaymentRepository defines the contract for payment data access
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id string) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
}
