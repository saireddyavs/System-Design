package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryCartRepo implements CartRepository with thread-safe in-memory storage
type InMemoryCartRepo struct {
	mu       sync.RWMutex
	carts    map[string]*models.Cart
	byUser   map[string]string
}

// NewInMemoryCartRepo creates a new in-memory cart repository
func NewInMemoryCartRepo() *InMemoryCartRepo {
	return &InMemoryCartRepo{
		carts:  make(map[string]*models.Cart),
		byUser: make(map[string]string),
	}
}

var _ interfaces.CartRepository = (*InMemoryCartRepo)(nil)

func (r *InMemoryCartRepo) Create(ctx context.Context, cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[cart.ID]; exists {
		return ErrAlreadyExists
	}
	cart.Items = make(map[string]models.CartItem)
	r.carts[cart.ID] = cart
	r.byUser[cart.UserID] = cart.ID
	return nil
}

func (r *InMemoryCartRepo) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.carts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyCart(c), nil
}

func (r *InMemoryCartRepo) GetByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byUser[userID]
	if !ok {
		return nil, ErrNotFound
	}
	return copyCart(r.carts[id]), nil
}

func (r *InMemoryCartRepo) Update(ctx context.Context, cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.carts[cart.ID]; !ok {
		return ErrNotFound
	}
	r.carts[cart.ID] = copyCart(cart)
	return nil
}

func (r *InMemoryCartRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.carts[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.byUser, c.UserID)
	delete(r.carts, id)
	return nil
}

func copyCart(c *models.Cart) *models.Cart {
	cp := *c
	cp.Items = make(map[string]models.CartItem)
	for k, v := range c.Items {
		cp.Items[k] = v
	}
	return &cp
}
