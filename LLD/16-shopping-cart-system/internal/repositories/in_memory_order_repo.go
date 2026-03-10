package repositories

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/models"
)

type InMemoryOrderRepository struct {
	mu       sync.RWMutex
	orders   map[string]*models.Order
	byUser   map[string][]string // userID -> orderIDs
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[string]*models.Order),
		byUser: make(map[string][]string),
	}
}

func (r *InMemoryOrderRepository) GetByID(id string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	return copyOrder(o), nil
}

func (r *InMemoryOrderRepository) GetByUserID(userID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids, ok := r.byUser[userID]
	if !ok {
		return []*models.Order{}, nil
	}
	result := make([]*models.Order, 0, len(ids))
	for _, id := range ids {
		if o, exists := r.orders[id]; exists {
			result = append(result, copyOrder(o))
		}
	}
	return result, nil
}

func (r *InMemoryOrderRepository) Create(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; exists {
		return fmt.Errorf("order already exists: %s", order.ID)
	}
	r.orders[order.ID] = copyOrder(order)
	r.byUser[order.UserID] = append(r.byUser[order.UserID], order.ID)
	return nil
}

func (r *InMemoryOrderRepository) Update(order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.orders[order.ID]; !exists {
		return fmt.Errorf("order not found: %s", order.ID)
	}
	r.orders[order.ID] = copyOrder(order)
	return nil
}

func copyOrder(o *models.Order) *models.Order {
	cpy := *o
	cpy.Items = make([]models.OrderItem, len(o.Items))
	copy(cpy.Items, o.Items)
	return &cpy
}
