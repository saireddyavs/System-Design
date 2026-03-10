package repositories

import (
	"context"
	"strings"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryProductRepo implements ProductRepository with thread-safe in-memory storage
type InMemoryProductRepo struct {
	mu       sync.RWMutex
	products map[string]*models.Product
	bySKU    map[string]string // SKU -> ProductID
}

// NewInMemoryProductRepo creates a new in-memory product repository
func NewInMemoryProductRepo() *InMemoryProductRepo {
	return &InMemoryProductRepo{
		products: make(map[string]*models.Product),
		bySKU:   make(map[string]string),
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
	r.bySKU[product.SKU] = product.ID
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

func (r *InMemoryProductRepo) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.bySKU[sku]
	if !ok {
		return nil, ErrNotFound
	}
	return copyProduct(r.products[id]), nil
}

func (r *InMemoryProductRepo) Update(ctx context.Context, product *models.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.products[product.ID]
	if !ok {
		return ErrNotFound
	}
	if existing.SKU != product.SKU {
		delete(r.bySKU, existing.SKU)
		r.bySKU[product.SKU] = product.ID
	}
	r.products[product.ID] = product
	return nil
}

func (r *InMemoryProductRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.products[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.bySKU, p.SKU)
	delete(r.products, id)
	return nil
}

func (r *InMemoryProductRepo) List(ctx context.Context, categoryID string, limit, offset int) ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Product
	for _, p := range r.products {
		if categoryID == "" || p.CategoryID == categoryID {
			result = append(result, copyProduct(p))
		}
	}
	return paginate(result, limit, offset), nil
}

func (r *InMemoryProductRepo) Search(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*models.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var result []*models.Product
	for _, p := range r.products {
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) && !strings.Contains(strings.ToLower(p.Description), query) {
			continue
		}
		if catID, ok := filters["category_id"].(string); ok && catID != "" && p.CategoryID != catID {
			continue
		}
		if minPrice, ok := filters["min_price"].(float64); ok && p.Price < minPrice {
			continue
		}
		if maxPrice, ok := filters["max_price"].(float64); ok && p.Price > maxPrice {
			continue
		}
		result = append(result, copyProduct(p))
	}
	return paginate(result, limit, offset), nil
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
