package repositories

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/models"
)

type InMemoryCartRepository struct {
	mu         sync.RWMutex
	carts      map[string]*models.Cart
	userCarts  map[string]string // userID -> cartID
	abandoned  map[string]*models.Cart // cartID -> cart (for abandoned tracking)
}

func NewInMemoryCartRepository() *InMemoryCartRepository {
	return &InMemoryCartRepository{
		carts:     make(map[string]*models.Cart),
		userCarts: make(map[string]string),
		abandoned: make(map[string]*models.Cart),
	}
}

func (r *InMemoryCartRepository) GetByID(id string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.carts[id]
	if !ok {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return copyCart(c), nil
}

func (r *InMemoryCartRepository) GetByUserID(userID string) (*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cartID, ok := r.userCarts[userID]
	if !ok {
		return nil, fmt.Errorf("no cart for user: %s", userID)
	}
	c, ok := r.carts[cartID]
	if !ok {
		return nil, fmt.Errorf("cart not found: %s", cartID)
	}
	// Return active cart only
	if c.Status != models.CartStatusActive {
		return nil, fmt.Errorf("user has no active cart: %s", userID)
	}
	return copyCart(c), nil
}

func (r *InMemoryCartRepository) Create(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.carts[cart.ID]; exists {
		return fmt.Errorf("cart already exists: %s", cart.ID)
	}
	r.carts[cart.ID] = copyCart(cart)
	r.userCarts[cart.UserID] = cart.ID
	return nil
}

func (r *InMemoryCartRepository) Update(cart *models.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.carts[cart.ID]; !exists {
		return fmt.Errorf("cart not found: %s", cart.ID)
	}
	r.carts[cart.ID] = copyCart(cart)
	return nil
}

func (r *InMemoryCartRepository) UpdateStatus(cartID string, status models.CartStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.carts[cartID]
	if !ok {
		return fmt.Errorf("cart not found: %s", cartID)
	}
	c.Status = status
	if status == models.CartStatusAbandoned {
		r.abandoned[cartID] = copyCart(c)
	}
	if status == models.CartStatusCheckedOut {
		delete(r.abandoned, cartID)
	}
	return nil
}

func (r *InMemoryCartRepository) GetAbandonedCarts(userID string) ([]*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Cart
	for _, c := range r.abandoned {
		if c.UserID == userID {
			result = append(result, copyCart(c))
		}
	}
	return result, nil
}

func (r *InMemoryCartRepository) GetAllAbandonedCarts() ([]*models.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Cart, 0, len(r.abandoned))
	for _, c := range r.abandoned {
		result = append(result, copyCart(c))
	}
	return result, nil
}

func copyCart(c *models.Cart) *models.Cart {
	cpy := *c
	cpy.Items = make([]models.CartItem, len(c.Items))
	copy(cpy.Items, c.Items)
	return &cpy
}
