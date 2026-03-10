package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"ecommerce-website/internal/factory"
	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/observer"
	"ecommerce-website/internal/repositories"
	"ecommerce-website/internal/services"
	"ecommerce-website/internal/strategies"
)

func setupOrderTest(t *testing.T) (context.Context, *services.OrderService, string) {
	ctx := context.Background()

	productRepo := repositories.NewInMemoryProductRepo()
	userRepo := repositories.NewInMemoryUserRepo()
	orderRepo := repositories.NewInMemoryOrderRepo()
	cartRepo := repositories.NewInMemoryCartRepo()
	paymentRepo := repositories.NewInMemoryPaymentRepo()
	couponRepo := repositories.NewInMemoryCouponRepo()

	var idCounter int64
	idGen := func() string {
		return fmt.Sprintf("id-%d", atomic.AddInt64(&idCounter, 1))
	}

	// Seed product
	prod := &models.Product{
		ID: "p1", Name: "Laptop", Price: 500, Stock: 10,
		CategoryID: "c1", SKU: "SKU-001",
	}
	_ = productRepo.Create(ctx, prod)

	// Seed user with address
	user := &models.User{
		ID: "user-1", Name: "John", Email: "john@test.com",
		Addresses: []models.Address{
			{ID: "a1", Street: "123 St", City: "NYC", Country: "USA", IsDefault: true},
		},
	}
	_ = userRepo.Create(ctx, user)

	// Seed coupon
	now := time.Now()
	coupon := &models.Coupon{
		ID: "cpn-1", Code: "SAVE10", Type: models.CouponTypePercentage, Value: 10,
		MinOrderAmount: 100, ExpiresAt: now.Add(24 * time.Hour), UsageLimit: 10,
		CreatedAt: now, UpdatedAt: now,
	}
	_ = couponRepo.Create(ctx, coupon)

	// Add to cart via cart service
	cartService := services.NewCartService(cartRepo, productRepo, idGen)
	_ = cartService.AddItem(ctx, "user-1", "p1", 2)

	orderObserver := observer.NewOrderStatusObserver()
	paymentService := services.NewPaymentService([]interfaces.PaymentProcessor{
		strategies.NewUPIProcessor(),
	})
	discountStrategies := []interfaces.DiscountStrategy{
		strategies.NewPercentageDiscountStrategy(),
		strategies.NewFlatDiscountStrategy(),
		strategies.NewBOGODiscountStrategy(),
	}
	couponService := services.NewCouponService(couponRepo, discountStrategies)
	orderFactory := factory.NewOrderFactory(idGen)

	orderService := services.NewOrderService(
		orderRepo, cartRepo, productRepo, paymentRepo, couponRepo,
		orderFactory, paymentService, couponService, orderObserver, idGen,
	)

	return ctx, orderService, "user-1"
}

func TestOrderService_PlaceOrder(t *testing.T) {
	ctx, orderService, userID := setupOrderTest(t)

	addr := models.Address{ID: "a1", Street: "123 St", City: "NYC", Country: "USA"}

	order, err := orderService.PlaceOrder(ctx, services.PlaceOrderInput{
		UserID:          userID,
		ShippingAddress: addr,
		CouponCode:      "SAVE10",
		PaymentMethod:   models.PaymentMethodUPI,
	})
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if order == nil {
		t.Fatal("order is nil")
	}
	if order.Status != models.OrderStatusConfirmed {
		t.Errorf("expected status Confirmed, got %s", order.Status)
	}
	// 2 * 500 = 1000, 10% = 100 discount, final = 900
	if order.FinalAmount != 900 {
		t.Errorf("expected final amount 900, got %.2f", order.FinalAmount)
	}
}

func TestOrderService_GetOrderHistory(t *testing.T) {
	ctx, orderService, userID := setupOrderTest(t)

	addr := models.Address{ID: "a1", Street: "123 St", City: "NYC", Country: "USA"}
	_, _ = orderService.PlaceOrder(ctx, services.PlaceOrderInput{
		UserID:          userID,
		ShippingAddress: addr,
		PaymentMethod:   models.PaymentMethodUPI,
	})

	history, err := orderService.GetOrderHistory(ctx, userID, 10, 0)
	if err != nil {
		t.Fatalf("GetOrderHistory failed: %v", err)
	}
	if len(history) < 1 {
		t.Errorf("expected at least 1 order in history, got %d", len(history))
	}
}
