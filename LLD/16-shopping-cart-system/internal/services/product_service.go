package services

import (
	"fmt"
	"time"

	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
)

type ProductService struct {
	repo interfaces.ProductRepository
}

func NewProductService(repo interfaces.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetByID(id string) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) GetAll() ([]*models.Product, error) {
	return s.repo.GetAll()
}

func (s *ProductService) CreateProduct(name, description, categoryID, sku string, price float64, stock int, weight float64) (*models.Product, error) {
	id := fmt.Sprintf("prod_%d", time.Now().UnixNano())
	product := &models.Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		CategoryID:  categoryID,
		Stock:       stock,
		SKU:         sku,
		Weight:      weight,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Create(product); err != nil {
		return nil, err
	}
	return product, nil
}
