package interfaces

import "shopping-cart-system/internal/models"

// OrderRepository defines data access for orders (Repository pattern)
type OrderRepository interface {
	GetByID(id string) (*models.Order, error)
	GetByUserID(userID string) ([]*models.Order, error)
	Create(order *models.Order) error
	Update(order *models.Order) error
}
