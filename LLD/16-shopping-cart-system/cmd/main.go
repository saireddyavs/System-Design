package main

import (
	"fmt"
	"log"
	"time"

	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/repositories"
	"shopping-cart-system/internal/services"
	"shopping-cart-system/internal/strategies"
)

func main() {
	// Repositories
	productRepo := repositories.NewInMemoryProductRepository()
	cartRepo := repositories.NewInMemoryCartRepository()
	orderRepo := repositories.NewInMemoryOrderRepository()
	couponRepo := repositories.NewInMemoryCouponRepository()
	userRepo := repositories.NewInMemoryUserRepository()

	// Seed data
	seedData(productRepo, couponRepo, userRepo)

	// Strategies
	discountRegistry := strategies.NewDiscountStrategyRegistry()
	taxCalculator := strategies.NewFlatTaxCalculator(0.18)
	paymentRegistry := strategies.NewPaymentProcessorRegistry()

	// Services
	cartService := services.NewCartService(cartRepo, productRepo)
	couponService := services.NewCouponService(couponRepo, discountRegistry)
	orderService := services.NewOrderService(orderRepo)

	// Register observer
	cartService.RegisterObserver(services.NewCartAbandonmentObserver(cartRepo))

	checkoutService := services.NewCheckoutService(
		cartRepo, orderRepo, productRepo, couponRepo, userRepo,
		cartService, couponService, orderService,
		taxCalculator, paymentRegistry,
	)

	// Demo flow
	userID := "user1"
	fmt.Println("=== Shopping Cart System Demo ===")

	// Add items to cart
	fmt.Println("1. Adding items to cart...")
	if err := cartService.AddItem(userID, "prod1", 2); err != nil {
		log.Fatalf("Add item failed: %v", err)
	}
	if err := cartService.AddItem(userID, "prod2", 1); err != nil {
		log.Fatalf("Add item failed: %v", err)
	}

	cart, _ := cartService.GetCart(userID)
	fmt.Printf("   Cart subtotal: $%.2f (%d items)\n\n", cart.Subtotal(), cart.ItemCount())

	// Checkout with coupon
	fmt.Println("2. Checkout with coupon SAVE10...")
	summary, err := checkoutService.Checkout(userID, "SAVE10", models.PaymentMethodCreditCard)
	if err != nil {
		log.Fatalf("Checkout failed: %v", err)
	}

	fmt.Printf("   Order ID: %s\n", summary.OrderID)
	fmt.Printf("   Subtotal: $%.2f\n", summary.Subtotal)
	fmt.Printf("   Discount: $%.2f\n", summary.Discount)
	fmt.Printf("   Tax: $%.2f\n", summary.Tax)
	fmt.Printf("   Total: $%.2f\n", summary.Total)
	fmt.Printf("   Payment ID: %s\n\n", summary.PaymentID)

	// Verify cart cleared
	cart, _ = cartService.GetCart(userID)
	fmt.Printf("3. Cart after checkout: %d items (empty)\n", len(cart.Items))

	// Order history
	orders, _ := orderService.GetByUserID(userID)
	fmt.Printf("\n4. Order history: %d order(s)\n", len(orders))
	for _, o := range orders {
		fmt.Printf("   - %s: $%.2f (%s)\n", o.ID, o.Total, o.Status)
	}

	fmt.Println("\n=== Demo Complete ===")
}

func seedData(productRepo *repositories.InMemoryProductRepository, couponRepo *repositories.InMemoryCouponRepository, userRepo *repositories.InMemoryUserRepository) {
	// Products
	products := []*models.Product{
		{ID: "prod1", Name: "Laptop", Description: "Gaming Laptop", Price: 999.99, CategoryID: "cat1", Stock: 10, SKU: "LAP-001", Weight: 2.5, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "prod2", Name: "Mouse", Description: "Wireless Mouse", Price: 29.99, CategoryID: "cat2", Stock: 50, SKU: "MOU-001", Weight: 0.2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "prod3", Name: "Keyboard", Description: "Mechanical Keyboard", Price: 89.99, CategoryID: "cat2", Stock: 30, SKU: "KEY-001", Weight: 0.8, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, p := range products {
		_ = productRepo.Create(p)
	}

	// Coupons
	coupons := []*models.Coupon{
		{ID: "c1", Code: "SAVE10", Type: models.CouponTypePercentage, Value: 10, MinOrderAmount: 50, ExpiresAt: time.Now().Add(30 * 24 * time.Hour), MaxUsageLimit: 100, CurrentUsage: 0, CreatedAt: time.Now()},
		{ID: "c2", Code: "FLAT20", Type: models.CouponTypeFlat, Value: 20, MinOrderAmount: 100, ExpiresAt: time.Now().Add(30 * 24 * time.Hour), MaxUsageLimit: 50, CurrentUsage: 0, CreatedAt: time.Now()},
		{ID: "c3", Code: "BOGO", Type: models.CouponTypeBOGO, Value: 1, MinOrderAmount: 0, ExpiresAt: time.Now().Add(30 * 24 * time.Hour), MaxUsageLimit: 0, CurrentUsage: 0, CreatedAt: time.Now()},
	}
	for _, c := range coupons {
		_ = couponRepo.Create(c)
	}

	// User
	user := &models.User{
		ID:    "user1",
		Name:  "John Doe",
		Email: "john@example.com",
		Address: models.Address{
			Street:     "123 Main St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "USA",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(user)
}
