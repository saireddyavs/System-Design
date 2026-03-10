package services

import (
	"context"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/observer"
	"ecommerce-website/internal/repositories"
)

// InventoryService manages stock and low stock alerts (Observer pattern)
type InventoryService struct {
	productRepo interfaces.ProductRepository
	observer    *observer.InventoryObserver
}

// NewInventoryService creates a new inventory service
func NewInventoryService(productRepo interfaces.ProductRepository, invObserver *observer.InventoryObserver) *InventoryService {
	return &InventoryService{
		productRepo: productRepo,
		observer:    invObserver,
	}
}

// DecrementStock reduces stock and triggers low stock alert if needed
func (s *InventoryService) DecrementStock(ctx context.Context, productID string, quantity int) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < quantity {
		return repositories.ErrInsufficientStock
	}

	if err := s.productRepo.DecrementStock(ctx, productID, quantity); err != nil {
		return err
	}

	// Fetch updated product for alert check
	updated, _ := s.productRepo.GetByID(ctx, productID)
	if updated != nil {
		s.observer.NotifyLowStock(ctx, productID, updated.Name, updated.Stock)
	}
	return nil
}

// IncrementStock increases stock (e.g., on return/cancel)
func (s *InventoryService) IncrementStock(ctx context.Context, productID string, quantity int) error {
	return s.productRepo.IncrementStock(ctx, productID, quantity)
}

// CheckStock returns current stock for a product
func (s *InventoryService) CheckStock(ctx context.Context, productID string) (int, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return 0, err
	}
	return product.Stock, nil
}

// CheckAndReserve validates stock availability (for order flow)
func (s *InventoryService) CheckAndReserve(ctx context.Context, productID string, quantity int) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Stock < quantity {
		return repositories.ErrInsufficientStock
	}
	return nil
}
