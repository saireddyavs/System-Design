package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// ProductRepository defines the contract for product data access (Repository pattern)
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	DecrementStock(ctx context.Context, productID string, quantity int) error
	IncrementStock(ctx context.Context, productID string, quantity int) error
}
