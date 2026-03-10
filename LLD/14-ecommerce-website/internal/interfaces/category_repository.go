package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// CategoryRepository defines the contract for category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	List(ctx context.Context, parentID *string) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
}
