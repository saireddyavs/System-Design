package interfaces

import (
	"food-delivery-system/internal/models"
	"time"
)

// OrderRepository defines the contract for order data access (Repository Pattern)
type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id string) (*models.Order, error)
	GetByCustomerID(customerID string) ([]*models.Order, error)
	GetByRestaurantID(restaurantID string) ([]*models.Order, error)
	GetByAgentID(agentID string) ([]*models.Order, error)
	Update(order *models.Order) error
	GetOrdersInStatus(status models.OrderStatus, since time.Time) ([]*models.Order, error)
}
