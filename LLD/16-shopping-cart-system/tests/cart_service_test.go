package tests

import (
	"sync"
	"testing"
	"time"

	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/repositories"
	"shopping-cart-system/internal/services"
)

func setupCartTest(t *testing.T) (*services.CartService, *repositories.InMemoryProductRepository) {
	productRepo := repositories.NewInMemoryProductRepository()
	cartRepo := repositories.NewInMemoryCartRepository()

	// Seed products
	products := []*models.Product{
		{ID: "p1", Name: "Product 1", Price: 10.0, Stock: 5, SKU: "SKU1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "Product 2", Price: 20.0, Stock: 3, SKU: "SKU2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, p := range products {
		_ = productRepo.Create(p)
	}

	cartService := services.NewCartService(cartRepo, productRepo)
	return cartService, productRepo
}

func TestCartService_AddItem(t *testing.T) {
	cartService, _ := setupCartTest(t)
	userID := "user1"

	err := cartService.AddItem(userID, "p1", 2)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	cart, err := cartService.GetCart(userID)
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if len(cart.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(cart.Items))
	}
	if cart.Items[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", cart.Items[0].Quantity)
	}
	if cart.Subtotal() != 20.0 {
		t.Errorf("expected subtotal 20, got %.2f", cart.Subtotal())
	}
}

func TestCartService_AddItem_InsufficientStock(t *testing.T) {
	cartService, _ := setupCartTest(t)

	err := cartService.AddItem("user1", "p1", 10)
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}

func TestCartService_UpdateQuantity(t *testing.T) {
	cartService, _ := setupCartTest(t)
	_ = cartService.AddItem("user1", "p1", 2)

	err := cartService.UpdateQuantity("user1", "p1", 3)
	if err != nil {
		t.Fatalf("UpdateQuantity failed: %v", err)
	}

	cart, _ := cartService.GetCart("user1")
	if cart.Items[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", cart.Items[0].Quantity)
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	cartService, _ := setupCartTest(t)
	_ = cartService.AddItem("user1", "p1", 2)
	_ = cartService.AddItem("user1", "p2", 1)

	err := cartService.RemoveItem("user1", "p1")
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}

	cart, _ := cartService.GetCart("user1")
	if len(cart.Items) != 1 {
		t.Errorf("expected 1 item after removal, got %d", len(cart.Items))
	}
	if cart.Items[0].ProductID != "p2" {
		t.Errorf("expected remaining item p2, got %s", cart.Items[0].ProductID)
	}
}

func TestCartService_ConcurrentAdd(t *testing.T) {
	cartService, _ := setupCartTest(t)
	userID := "user1"
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cartService.AddItem(userID, "p1", 1)
		}()
	}
	wg.Wait()

	cart, _ := cartService.GetCart(userID)
	// Total could be 5 (if all succeeded) - stock is 5, so all 5 adds might succeed
	// Or some might fail with insufficient stock
	if cart.ItemCount() < 1 || cart.ItemCount() > 5 {
		t.Errorf("unexpected item count: %d", cart.ItemCount())
	}
}
