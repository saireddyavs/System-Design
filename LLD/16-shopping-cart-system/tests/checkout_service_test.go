package tests

import (
	"testing"
	"time"

	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/repositories"
	"shopping-cart-system/internal/services"
	"shopping-cart-system/internal/strategies"
)

func setupCheckoutTest(t *testing.T) (*services.CheckoutService, string) {
	productRepo := repositories.NewInMemoryProductRepository()
	cartRepo := repositories.NewInMemoryCartRepository()
	orderRepo := repositories.NewInMemoryOrderRepository()
	couponRepo := repositories.NewInMemoryCouponRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	// User
	user := &models.User{
		ID:    "user1",
		Name:  "Test User",
		Email: "test@test.com",
		Address: models.Address{Street: "123 St", City: "NY", State: "NY", PostalCode: "10001", Country: "USA"},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(user)

	// Products
	products := []*models.Product{
		{ID: "p1", Name: "Item 1", Price: 100, Stock: 10, SKU: "S1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "p2", Name: "Item 2", Price: 50, Stock: 5, SKU: "S2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, p := range products {
		_ = productRepo.Create(p)
	}

	// Coupon: 10% off, min $50
	_ = couponRepo.Create(&models.Coupon{
		ID: "c1", Code: "SAVE10", Type: models.CouponTypePercentage, Value: 10,
		MinOrderAmount: 50, ExpiresAt: time.Now().Add(24 * time.Hour), MaxUsageLimit: 10, CurrentUsage: 0, CreatedAt: time.Now(),
	})

	cartService := services.NewCartService(cartRepo, productRepo)
	couponService := services.NewCouponService(couponRepo, strategies.NewDiscountStrategyRegistry())
	orderService := services.NewOrderService(orderRepo)
	taxCalc := strategies.NewFlatTaxCalculator(0.18)
	paymentReg := strategies.NewPaymentProcessorRegistry()

	checkoutService := services.NewCheckoutService(
		cartRepo, orderRepo, productRepo, couponRepo, userRepo,
		cartService, couponService, orderService, taxCalc, paymentReg,
	)

	// Add items to cart
	_ = cartService.AddItem("user1", "p1", 2) // 200
	_ = cartService.AddItem("user1", "p2", 1) // 50 -> subtotal 250

	return checkoutService, "user1"
}

func TestCheckoutService_Checkout_Success(t *testing.T) {
	checkoutService, userID := setupCheckoutTest(t)

	summary, err := checkoutService.Checkout(userID, "SAVE10", models.PaymentMethodCreditCard)
	if err != nil {
		t.Fatalf("Checkout failed: %v", err)
	}

	// Subtotal 250, 10% discount = 25, after discount = 225, tax 18% = 40.5, total = 265.5
	if summary.Subtotal != 250 {
		t.Errorf("expected subtotal 250, got %.2f", summary.Subtotal)
	}
	if summary.Discount != 25 {
		t.Errorf("expected discount 25, got %.2f", summary.Discount)
	}
	expectedTax := 225 * 0.18
	if summary.Tax != expectedTax {
		t.Errorf("expected tax %.2f, got %.2f", expectedTax, summary.Tax)
	}
	if summary.OrderID == "" {
		t.Error("expected order ID")
	}
	if summary.PaymentID == "" {
		t.Error("expected payment ID")
	}
}

func TestCheckoutService_Checkout_EmptyCart(t *testing.T) {
	productRepo := repositories.NewInMemoryProductRepository()
	cartRepo := repositories.NewInMemoryCartRepository()
	orderRepo := repositories.NewInMemoryOrderRepository()
	couponRepo := repositories.NewInMemoryCouponRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	user := &models.User{ID: "u1", Name: "U", Email: "u@u.com", Address: models.Address{}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = userRepo.Create(user)

	cartService := services.NewCartService(cartRepo, productRepo)
	couponService := services.NewCouponService(couponRepo, strategies.NewDiscountStrategyRegistry())
	orderService := services.NewOrderService(orderRepo)

	checkoutService := services.NewCheckoutService(
		cartRepo, orderRepo, productRepo, couponRepo, userRepo,
		cartService, couponService, orderService,
		strategies.NewFlatTaxCalculator(0.18), strategies.NewPaymentProcessorRegistry(),
	)

	_, err := checkoutService.Checkout("u1", "", models.PaymentMethodCreditCard)
	if err == nil {
		t.Fatal("expected error for empty cart")
	}
}

func TestCheckoutService_Checkout_StockValidation(t *testing.T) {
	productRepo := repositories.NewInMemoryProductRepository()
	cartRepo := repositories.NewInMemoryCartRepository()
	orderRepo := repositories.NewInMemoryOrderRepository()
	couponRepo := repositories.NewInMemoryCouponRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	user := &models.User{ID: "u1", Name: "U", Email: "u@u.com", Address: models.Address{}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = userRepo.Create(user)
	// Stock 2, we add 2 items. Then simulate another order taking stock.
	_ = productRepo.Create(&models.Product{ID: "p1", Name: "P", Price: 10, Stock: 2, SKU: "S", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	cartService := services.NewCartService(cartRepo, productRepo)
	_ = cartService.AddItem("u1", "p1", 2)

	// Simulate another order depleting stock - now only 1 left, cart needs 2
	_ = productRepo.DecrementStock("p1", 1)

	couponService := services.NewCouponService(couponRepo, strategies.NewDiscountStrategyRegistry())
	orderService := services.NewOrderService(orderRepo)

	checkoutService := services.NewCheckoutService(
		cartRepo, orderRepo, productRepo, couponRepo, userRepo,
		cartService, couponService, orderService,
		strategies.NewFlatTaxCalculator(0.18), strategies.NewPaymentProcessorRegistry(),
	)

	_, err := checkoutService.Checkout("u1", "", models.PaymentMethodCreditCard)
	if err == nil {
		t.Fatal("expected error for insufficient stock during checkout")
	}
}
