package interfaces

import "online-bookstore/internal/models"

// OrderRepository defines the contract for order data access.
// ISP: Separate from BookRepository - single responsibility per interface.
type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id string) (*models.Order, error)
	GetByUserID(userID string) ([]*models.Order, error)
	Update(order *models.Order) error
}
