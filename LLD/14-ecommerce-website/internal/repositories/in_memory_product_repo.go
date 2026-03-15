package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryProductRepo implements ProductRepository with thread-safe in-memory storage
type InMemoryProductRepo struct {
	mu       sync.RWMutex
	products map[string]*models.Product
}

// NewInMemoryProductRepo creates a new in-memory product repository
func NewInMemoryProductRepo() *InMemoryProductRepo {
	return &InMemoryProductRepo{
		products: make(map[string]*models.Product),
	}
}

// Ensure implementation
var _ interfaces.ProductRepository = (*InMemoryProductRepo)(nil)

func (r *InMemoryProductRepo) Create(ctx context.Context, product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; exists {
		return ErrAlreadyExists
	}
	r.products[product.ID] = product
	return nil
}

func (r *InMemoryProductRepo) GetByID(ctx context.Context, id string) (*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.products[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyProduct(p), nil
}

func (r *InMemoryProductRepo) DecrementStock(ctx context.Context, productID string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.products[productID]
	if !ok {
		return ErrNotFound
	}
	if p.Stock < quantity {
		return ErrInsufficientStock
	}
	p.Stock -= quantity
	return nil
}

func (r *InMemoryProductRepo) IncrementStock(ctx context.Context, productID string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.products[productID]
	if !ok {
		return ErrNotFound
	}
	p.Stock += quantity
	return nil
}

func copyProduct(p *models.Product) *models.Product {
	cp := *p
	cp.Images = make([]string, len(p.Images))
	copy(cp.Images, p.Images)
	return &cp
}
