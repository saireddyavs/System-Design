package tests

import (
	"context"
	"sync"
	"testing"

	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
)

func TestConcurrentStockDecrement(t *testing.T) {
	ctx := context.Background()
	repo := repositories.NewInMemoryProductRepo()

	// Product with 100 stock
	prod := &models.Product{
		ID: "p1", Name: "Concurrent Product", Stock: 100,
		CategoryID: "c1", SKU: "CONC-001",
	}
	_ = repo.Create(ctx, prod)

	// 20 goroutines each decrement 5 = 100 total
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = repo.DecrementStock(ctx, "p1", 5)
		}()
	}
	wg.Wait()

	updated, _ := repo.GetByID(ctx, "p1")
	if updated.Stock != 0 {
		t.Errorf("expected stock 0 after 20*5 decrements, got %d", updated.Stock)
	}
}

func TestConcurrentStockDecrement_InsufficientStock(t *testing.T) {
	ctx := context.Background()
	repo := repositories.NewInMemoryProductRepo()

	prod := &models.Product{
		ID: "p1", Name: "Low Stock", Stock: 10,
		CategoryID: "c1", SKU: "LOW-001",
	}
	_ = repo.Create(ctx, prod)

	// 5 goroutines each try to decrement 5 - only 2 should succeed fully
	// Actually with 10 stock, 2 can get 5 each. The rest will get ErrInsufficientStock
	errCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.DecrementStock(ctx, "p1", 5)
			if err != nil {
				mu.Lock()
				errCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	updated, _ := repo.GetByID(ctx, "p1")
	// 2 successful decrements of 5 = 10, stock = 0. 3 should fail.
	if updated.Stock != 0 {
		t.Errorf("expected stock 0, got %d", updated.Stock)
	}
	if errCount != 3 {
		t.Errorf("expected 3 insufficient stock errors, got %d", errCount)
	}
}

func TestConcurrentCartUpdates(t *testing.T) {
	ctx := context.Background()
	productRepo := repositories.NewInMemoryProductRepo()
	cartRepo := repositories.NewInMemoryCartRepo()

	prod := &models.Product{
		ID: "p1", Name: "Product", Stock: 1000, Price: 10,
		CategoryID: "c1", SKU: "P-001",
	}
	_ = productRepo.Create(ctx, prod)

	cart := &models.Cart{
		ID:     "cart-1",
		UserID: "user-1",
		Items:  make(map[string]models.CartItem),
	}
	_ = cartRepo.Create(ctx, cart)

	// Multiple goroutines add to cart concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, _ := cartRepo.GetByUserID(ctx, "user-1")
			if c != nil {
				item, ok := c.Items["p1"]
				if !ok {
					item = models.CartItem{ProductID: "p1", Quantity: 0, Price: 10}
				}
				item.Quantity += 1
				c.Items["p1"] = item
				_ = cartRepo.Update(ctx, c)
			}
		}()
	}
	wg.Wait()

	// Final cart should have some items (race condition possible, but no panic)
	final, _ := cartRepo.GetByUserID(ctx, "user-1")
	if final == nil {
		t.Fatal("cart is nil")
	}
	// At least 1 item should have been added (we may lose some to race)
	if len(final.Items) == 0 {
		t.Error("expected at least 1 item in cart")
	}
}
