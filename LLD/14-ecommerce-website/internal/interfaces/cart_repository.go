package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// CartRepository defines the contract for cart data access (Repository pattern)
type CartRepository interface {
	Create(ctx context.Context, cart *models.Cart) error
	GetByID(ctx context.Context, id string) (*models.Cart, error)
	GetByUserID(ctx context.Context, userID string) (*models.Cart, error)
	Update(ctx context.Context, cart *models.Cart) error
	Delete(ctx context.Context, id string) error
}
