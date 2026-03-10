package tests

import (
	"context"
	"testing"

	"ecommerce-website/internal/models"
	"ecommerce-website/internal/observer"
	"ecommerce-website/internal/repositories"
	"ecommerce-website/internal/services"
)

func setupInventoryTest(t *testing.T) (context.Context, *services.InventoryService, *repositories.InMemoryProductRepo) {
	ctx := context.Background()
	productRepo := repositories.NewInMemoryProductRepo()
	invObserver := observer.NewInventoryObserver(5)

	prod := &models.Product{
		ID: "p1", Name: "Widget", Stock: 10, CategoryID: "c1", SKU: "W-001",
	}
	_ = productRepo.Create(ctx, prod)

	inventoryService := services.NewInventoryService(productRepo, invObserver)
	return ctx, inventoryService, productRepo
}

func TestInventoryService_DecrementStock(t *testing.T) {
	ctx, invService, productRepo := setupInventoryTest(t)

	err := invService.DecrementStock(ctx, "p1", 3)
	if err != nil {
		t.Fatalf("DecrementStock failed: %v", err)
	}

	stock, err := invService.CheckStock(ctx, "p1")
	if err != nil {
		t.Fatalf("CheckStock failed: %v", err)
	}
	if stock != 7 {
		t.Errorf("expected stock 7, got %d", stock)
	}

	_ = productRepo
}

func TestInventoryService_DecrementStock_Insufficient(t *testing.T) {
	ctx, invService, _ := setupInventoryTest(t)

	err := invService.DecrementStock(ctx, "p1", 100)
	if err != repositories.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestInventoryService_IncrementStock(t *testing.T) {
	ctx, invService, _ := setupInventoryTest(t)

	_ = invService.DecrementStock(ctx, "p1", 3)
	err := invService.IncrementStock(ctx, "p1", 2)
	if err != nil {
		t.Fatalf("IncrementStock failed: %v", err)
	}

	stock, _ := invService.CheckStock(ctx, "p1")
	if stock != 9 {
		t.Errorf("expected stock 9 (10-3+2), got %d", stock)
	}
}

func TestInventoryService_CheckAndReserve(t *testing.T) {
	ctx, invService, _ := setupInventoryTest(t)

	err := invService.CheckAndReserve(ctx, "p1", 5)
	if err != nil {
		t.Errorf("CheckAndReserve should succeed, got %v", err)
	}

	err = invService.CheckAndReserve(ctx, "p1", 100)
	if err != repositories.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}
