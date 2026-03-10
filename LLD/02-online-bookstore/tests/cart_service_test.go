package tests

import (
	"testing"
	"time"

	"online-bookstore/internal/models"
	"online-bookstore/internal/repositories"
	"online-bookstore/internal/services"
)

func setupCartTest(t *testing.T) (*services.CartService, *models.Book, string) {
	bookRepo := repositories.NewInMemoryBookRepository()
	cartRepo := repositories.NewInMemoryCartRepository()

	book := &models.Book{
		ID:        "book-1",
		Title:     "Test Book",
		Author:    "Author",
		ISBN:      "978-123",
		Price:     19.99,
		Genre:     "Fiction",
		Stock:     10,
		CreatedAt: time.Now(),
	}
	_ = bookRepo.Create(book)

	cartSvc := services.NewCartService(cartRepo, bookRepo)
	userID := "user-1"
	cart := &models.Cart{
		ID:        "cart-1",
		UserID:    userID,
		Items:     make(map[string]int),
		UpdatedAt: time.Now(),
	}
	_ = cartRepo.Create(cart)

	return cartSvc, book, userID
}

func TestCartService_AddToCart(t *testing.T) {
	cartSvc, book, userID := setupCartTest(t)

	err := cartSvc.AddToCart(userID, book.ID, 3)
	if err != nil {
		t.Fatalf("AddToCart failed: %v", err)
	}

	cart, err := cartSvc.GetCart(userID)
	if err != nil {
		t.Fatalf("GetCart failed: %v", err)
	}
	if cart.Items[book.ID] != 3 {
		t.Errorf("expected quantity 3, got %d", cart.Items[book.ID])
	}
}

func TestCartService_AddToCart_InsufficientStock(t *testing.T) {
	cartSvc, book, userID := setupCartTest(t)

	err := cartSvc.AddToCart(userID, book.ID, 100)
	if err != services.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestCartService_RemoveFromCart(t *testing.T) {
	cartSvc, book, userID := setupCartTest(t)
	_ = cartSvc.AddToCart(userID, book.ID, 2)

	err := cartSvc.RemoveFromCart(userID, book.ID)
	if err != nil {
		t.Fatalf("RemoveFromCart failed: %v", err)
	}

	cart, _ := cartSvc.GetCart(userID)
	if _, exists := cart.Items[book.ID]; exists {
		t.Error("expected item to be removed")
	}
}

func TestCartService_UpdateQuantity(t *testing.T) {
	cartSvc, book, userID := setupCartTest(t)
	_ = cartSvc.AddToCart(userID, book.ID, 2)

	err := cartSvc.UpdateQuantity(userID, book.ID, 5)
	if err != nil {
		t.Fatalf("UpdateQuantity failed: %v", err)
	}

	cart, _ := cartSvc.GetCart(userID)
	if cart.Items[book.ID] != 5 {
		t.Errorf("expected quantity 5, got %d", cart.Items[book.ID])
	}
}
