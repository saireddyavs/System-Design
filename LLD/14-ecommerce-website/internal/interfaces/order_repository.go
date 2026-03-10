package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// OrderRepository defines the contract for order data access (Repository pattern)
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) error
}
