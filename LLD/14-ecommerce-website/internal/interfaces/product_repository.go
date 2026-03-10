package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// ProductRepository defines the contract for product data access (Repository pattern)
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, categoryID string, limit, offset int) ([]*models.Product, error)
	Search(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*models.Product, error)
	DecrementStock(ctx context.Context, productID string, quantity int) error
	IncrementStock(ctx context.Context, productID string, quantity int) error
}
