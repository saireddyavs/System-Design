package repositories

import (
	"errors"
	"sync"

	"online-bookstore/internal/models"
)

var ErrCartNotFound = errors.New("cart not found")

// InMemoryCartRepository implements CartRepository with thread-safe in-memory storage.
type InMemoryCartRepository struct {
	carts map[string]*models.Cart
	mu    sync.RWMutex
}

// NewInMemoryCartRepository creates a new in-memory cart repository.
func NewInMemoryCartRepository() *InMemoryCartRepository {
	return &InMemoryCartRepository{
		carts: make(map[string]*models.Cart),
	}
}

func (r *InMemoryCartRepository) Create(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[cart.ID]; exists {
		return errors.New("cart already exists")
	}
	r.carts[cart.ID] = cart
	return nil
}

func (r *InMemoryCartRepository) GetByID(id string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.carts[id]
	if !exists {
		return nil, ErrCartNotFound
	}
	return cart, nil
}

func (r *InMemoryCartRepository) GetByUserID(userID string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cart := range r.carts {
		if cart.UserID == userID {
			return cart, nil
		}
	}
	return nil, ErrCartNotFound
}

func (r *InMemoryCartRepository) Update(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[cart.ID]; !exists {
		return ErrCartNotFound
	}
	r.carts[cart.ID] = cart
	return nil
}

func (r *InMemoryCartRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[id]; !exists {
		return ErrCartNotFound
	}
	delete(r.carts, id)
	return nil
}
