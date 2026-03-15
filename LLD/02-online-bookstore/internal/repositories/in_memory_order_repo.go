package repositories

import (
	"errors"
	"sync"

	"online-bookstore/internal/models"
)

var ErrOrderNotFound = errors.New("order not found")

// InMemoryOrderRepository implements OrderRepository with thread-safe in-memory storage.
type InMemoryOrderRepository struct {
	orders map[string]*models.Order
	mu     sync.RWMutex
}

// NewInMemoryOrderRepository creates a new in-memory order repository.
func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*models.Order),
	}
}

func (r *InMemoryOrderRepository) Create(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}
	r.orders[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) GetByID(id string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (r *InMemoryOrderRepository) GetByUserID(userID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepository) Update(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return ErrOrderNotFound
	}
	r.orders[order.ID] = order
	return nil
}
