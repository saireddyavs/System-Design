package repositories

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/models"
)

type InMemoryProductRepository struct {
	mu       sync.RWMutex
	products map[string]*models.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: make(map[string]*models.Product),
	}
}

func (r *InMemoryProductRepository) GetByID(id string) (*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok {
		return nil, fmt.Errorf("product not found: %s", id)
	}
	// Return copy to prevent external mutation
	return copyProduct(p), nil
}

func (r *InMemoryProductRepository) GetByIDs(ids []string) ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := r.products[id]; ok {
			result = append(result, copyProduct(p))
		}
	}
	return result, nil
}

func (r *InMemoryProductRepository) GetAll() ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Product, 0, len(r.products))
	for _, p := range r.products {
		result = append(result, copyProduct(p))
	}
	return result, nil
}

func (r *InMemoryProductRepository) Create(product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.products[product.ID]; exists {
		return fmt.Errorf("product already exists: %s", product.ID)
	}
	r.products[product.ID] = copyProduct(product)
	return nil
}

func (r *InMemoryProductRepository) Update(product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.products[product.ID]; !exists {
		return fmt.Errorf("product not found: %s", product.ID)
	}
	r.products[product.ID] = copyProduct(product)
	return nil
}

func (r *InMemoryProductRepository) DecrementStock(productID string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[productID]
	if !ok {
		return fmt.Errorf("product not found: %s", productID)
	}
	if p.Stock < quantity {
		return fmt.Errorf("insufficient stock for product %s: have %d, need %d", productID, p.Stock, quantity)
	}
	p.Stock -= quantity
	return nil
}

func copyProduct(p *models.Product) *models.Product {
	cpy := *p
	return &cpy
}
