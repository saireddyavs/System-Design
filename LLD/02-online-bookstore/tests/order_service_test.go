package tests

import (
	"testing"
	"time"

	"online-bookstore/internal/models"
	"online-bookstore/internal/repositories"
	"online-bookstore/internal/services"
	"online-bookstore/internal/strategies"
)

func setupOrderTest(t *testing.T) (*services.OrderService, *models.Book, string) {
	bookRepo := repositories.NewInMemoryBookRepository()
	orderRepo := repositories.NewInMemoryOrderRepository()
	cartRepo := repositories.NewInMemoryCartRepository()

	book := &models.Book{
		ID:        "book-order-1",
		Title:     "Order Test Book",
		Author:    "Author",
		ISBN:      "978-order-1",
		Price:     25.00,
		Genre:     "Tech",
		Stock:     10,
		CreatedAt: time.Now(),
	}
	_ = bookRepo.Create(book)

	userID := "user-order-1"
	cart := &models.Cart{
		ID:        "cart-order-1",
		UserID:    userID,
		Items:     map[string]int{book.ID: 2},
		UpdatedAt: time.Now(),
	}
	_ = cartRepo.Create(cart)

	orderFactory := strategies.NewOrderFactory()
	paymentReg := strategies.NewPaymentProcessorRegistry()
	orderSvc := services.NewOrderService(orderRepo, cartRepo, bookRepo, orderFactory, paymentReg)

	return orderSvc, book, userID
}

func TestOrderService_PlaceOrder(t *testing.T) {
	orderSvc, _, userID := setupOrderTest(t)

	order, err := orderSvc.PlaceOrder(userID, "credit_card")
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if order.ID == "" {
		t.Error("expected non-empty order ID")
	}
	if order.TotalAmount != 50.00 {
		t.Errorf("expected total 50.00, got %.2f", order.TotalAmount)
	}
	if order.Status != models.OrderStatusPaid {
		t.Errorf("expected status paid, got %s", order.Status)
	}
}

func TestOrderService_PlaceOrder_InvalidPaymentMethod(t *testing.T) {
	orderSvc, _, userID := setupOrderTest(t)

	_, err := orderSvc.PlaceOrder(userID, "invalid_method")
	if err != services.ErrInvalidPaymentMethod {
		t.Errorf("expected ErrInvalidPaymentMethod, got %v", err)
	}
}

func TestOrderService_GetOrderHistory(t *testing.T) {
	orderSvc, _, userID := setupOrderTest(t)
	_, _ = orderSvc.PlaceOrder(userID, "credit_card")

	history, err := orderSvc.GetOrderHistory(userID)
	if err != nil {
		t.Fatalf("GetOrderHistory failed: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 order in history, got %d", len(history))
	}
}
