package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
	"ecommerce-website/internal/services"
)

func setupCartTest(t *testing.T) (context.Context, *services.CartService, *repositories.InMemoryProductRepo) {
	ctx := context.Background()
	productRepo := repositories.NewInMemoryProductRepo()
	cartRepo := repositories.NewInMemoryCartRepo()

	var idCounter int64
	idGen := func() string {
		return fmt.Sprintf("cart-%d", atomic.AddInt64(&idCounter, 1))
	}

	// Seed product
	prod := &models.Product{
		ID: "p1", Name: "Test Product", Price: 99.99, Stock: 10,
		CategoryID: "c1", SKU: "SKU-001",
	}
	_ = productRepo.Create(ctx, prod)

	cartService := services.NewCartService(cartRepo, productRepo, idGen)
	return ctx, cartService, productRepo
}

func TestCartService_AddItem(t *testing.T) {
	ctx, cartService, _ := setupCartTest(t)

	err := cartService.AddItem(ctx, "user-1", "p1", 2)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	cart, err := cartService.GetCart(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(cart.Items))
	}
	item, ok := cart.Items["p1"]
	if !ok {
		t.Fatal("item p1 not in cart")
	}
	if item.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", item.Quantity)
	}
	if item.Price != 99.99 {
		t.Errorf("expected price 99.99, got %.2f", item.Price)
	}
}

func TestCartService_AddItem_InsufficientStock(t *testing.T) {
	ctx, cartService, _ := setupCartTest(t)

	err := cartService.AddItem(ctx, "user-1", "p1", 100)
	if err != repositories.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestCartService_UpdateQuantity(t *testing.T) {
	ctx, cartService, _ := setupCartTest(t)

	_ = cartService.AddItem(ctx, "user-1", "p1", 2)
	err := cartService.UpdateQuantity(ctx, "user-1", "p1", 5)
	if err != nil {
		t.Fatalf("UpdateQuantity failed: %v", err)
	}

	cart, _ := cartService.GetCart(ctx, "user-1")
	if cart.Items["p1"].Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", cart.Items["p1"].Quantity)
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	ctx, cartService, _ := setupCartTest(t)

	_ = cartService.AddItem(ctx, "user-1", "p1", 2)
	err := cartService.RemoveItem(ctx, "user-1", "p1")
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}

	cart, _ := cartService.GetCart(ctx, "user-1")
	if len(cart.Items) != 0 {
		t.Errorf("expected 0 items after remove, got %d", len(cart.Items))
	}
}

func TestCartService_ClearCart(t *testing.T) {
	ctx, cartService, _ := setupCartTest(t)

	_ = cartService.AddItem(ctx, "user-1", "p1", 2)
	err := cartService.ClearCart(ctx, "user-1")
	if err != nil {
		t.Fatalf("ClearCart failed: %v", err)
	}

	cart, _ := cartService.GetCart(ctx, "user-1")
	if len(cart.Items) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(cart.Items))
	}
}
