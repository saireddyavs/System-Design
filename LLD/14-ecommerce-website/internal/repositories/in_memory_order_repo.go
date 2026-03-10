package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryOrderRepo implements OrderRepository with thread-safe in-memory storage
type InMemoryOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

// NewInMemoryOrderRepo creates a new in-memory order repository
func NewInMemoryOrderRepo() *InMemoryOrderRepo {
	return &InMemoryOrderRepo{
		orders: make(map[string]*models.Order),
	}
}

var _ interfaces.OrderRepository = (*InMemoryOrderRepo)(nil)

func (r *InMemoryOrderRepo) Create(ctx context.Context, order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return ErrAlreadyExists
	}
	r.orders[order.ID] = copyOrder(order)
	return nil
}

func (r *InMemoryOrderRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyOrder(o), nil
}

func (r *InMemoryOrderRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Order
	for _, o := range r.orders {
		if o.UserID == userID {
			result = append(result, copyOrder(o))
		}
	}
	return paginateOrders(result, limit, offset), nil
}

func (r *InMemoryOrderRepo) Update(ctx context.Context, order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.orders[order.ID]; !ok {
		return ErrNotFound
	}
	r.orders[order.ID] = copyOrder(order)
	return nil
}

func (r *InMemoryOrderRepo) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[orderID]
	if !ok {
		return ErrNotFound
	}
	o.Status = status
	return nil
}

func copyOrder(o *models.Order) *models.Order {
	cp := *o
	cp.Items = make([]models.OrderItem, len(o.Items))
	copy(cp.Items, o.Items)
	return &cp
}

func paginateOrders(orders []*models.Order, limit, offset int) []*models.Order {
	if offset >= len(orders) {
		return nil
	}
	end := offset + limit
	if end > len(orders) {
		end = len(orders)
	}
	return orders[offset:end]
}
