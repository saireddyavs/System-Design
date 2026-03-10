package services

import (
	"context"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// ProductService handles product catalog operations
type ProductService struct {
	repo interfaces.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo interfaces.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create adds a new product to the catalog
func (s *ProductService) Create(ctx context.Context, product *models.Product) error {
	return s.repo.Create(ctx, product)
}

// GetByID retrieves a product by ID
func (s *ProductService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySKU retrieves a product by SKU
func (s *ProductService) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	return s.repo.GetBySKU(ctx, sku)
}

// Update updates an existing product
func (s *ProductService) Update(ctx context.Context, product *models.Product) error {
	return s.repo.Update(ctx, product)
}

// Delete removes a product
func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List returns products, optionally filtered by category
func (s *ProductService) List(ctx context.Context, categoryID string, limit, offset int) ([]*models.Product, error) {
	return s.repo.List(ctx, categoryID, limit, offset)
}

// Search searches products with filters
func (s *ProductService) Search(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*models.Product, error) {
	return s.repo.Search(ctx, query, filters, limit, offset)
}
