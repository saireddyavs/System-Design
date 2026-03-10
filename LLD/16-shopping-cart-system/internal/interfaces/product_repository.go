package interfaces

import "shopping-cart-system/internal/models"

// ProductRepository defines data access for products (Repository pattern)
type ProductRepository interface {
	GetByID(id string) (*models.Product, error)
	GetByIDs(ids []string) ([]*models.Product, error)
	GetAll() ([]*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	DecrementStock(productID string, quantity int) error
}
