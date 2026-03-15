package repositories

import (
	"errors"
	"food-delivery-system/internal/models"
	"sync"
)

var ErrOrderNotFound = errors.New("order not found")

// InMemoryOrderRepo implements OrderRepository with thread-safe in-memory storage
type InMemoryOrderRepo struct {
	orders map[string]*models.Order
	mu     sync.RWMutex
}

// NewInMemoryOrderRepo creates a new in-memory order repository
func NewInMemoryOrderRepo() *InMemoryOrderRepo {
	return &InMemoryOrderRepo{
		orders: make(map[string]*models.Order),
	}
}

func (r *InMemoryOrderRepo) Create(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepo) GetByID(id string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (r *InMemoryOrderRepo) GetByCustomerID(customerID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Order
	for _, o := range r.orders {
		if o.CustomerID == customerID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepo) GetByRestaurantID(restaurantID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Order
	for _, o := range r.orders {
		if o.RestaurantID == restaurantID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepo) GetByAgentID(agentID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Order
	for _, o := range r.orders {
		if o.AgentID == agentID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepo) Update(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; !exists {
		return ErrOrderNotFound
	}
	r.orders[order.ID] = order
	return nil
}
