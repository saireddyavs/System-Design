package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ecommerce-website/internal/factory"
	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/observer"
	"ecommerce-website/internal/repositories"
	"ecommerce-website/internal/services"
	"ecommerce-website/internal/strategies"
)

func main() {
	ctx := context.Background()

	// Repositories
	productRepo := repositories.NewInMemoryProductRepo()
	userRepo := repositories.NewInMemoryUserRepo()
	orderRepo := repositories.NewInMemoryOrderRepo()
	cartRepo := repositories.NewInMemoryCartRepo()
	paymentRepo := repositories.NewInMemoryPaymentRepo()
	couponRepo := repositories.NewInMemoryCouponRepo()
	categoryRepo := repositories.NewInMemoryCategoryRepo()

	// ID generator
	idGen := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Observers
	orderObserver := observer.NewOrderStatusObserver()
	notifier := services.NewLoggingNotificationService()
	orderObserver.Subscribe(notifier)

	// Payment processors (Strategy pattern)
	paymentService := services.NewPaymentService([]interfaces.PaymentProcessor{
		strategies.NewCreditCardProcessor(),
		strategies.NewDebitCardProcessor(),
		strategies.NewUPIProcessor(),
		strategies.NewWalletProcessor(),
	})

	// Discount strategies
	discountStrategies := []interfaces.DiscountStrategy{
		strategies.NewPercentageDiscountStrategy(),
		strategies.NewFlatDiscountStrategy(),
		strategies.NewBOGODiscountStrategy(),
	}

	// Services
	couponService := services.NewCouponService(couponRepo, discountStrategies)
	orderFactory := factory.NewOrderFactory(idGen)
	orderService := services.NewOrderService(
		orderRepo, cartRepo, productRepo, paymentRepo, couponRepo,
		orderFactory, paymentService, couponService, orderObserver, idGen,
	)
	cartService := services.NewCartService(cartRepo, productRepo, idGen)
	userService := services.NewUserService(userRepo)

	// Seed data
	seedData(ctx, productRepo, categoryRepo, userRepo, couponRepo)

	// Demo flow
	user, _ := userService.GetByEmail(ctx, "john@example.com")
	if user == nil {
		log.Fatal("User not found")
	}

	// Add to cart
	_ = cartService.AddItem(ctx, user.ID, "prod-1", 2)
	cart, _ := cartService.GetCart(ctx, user.ID)
	fmt.Printf("Cart has %d items\n", len(cart.Items))

	// Place order
	addr := user.Addresses[0]
	order, err := orderService.PlaceOrder(ctx, services.PlaceOrderInput{
		UserID:          user.ID,
		ShippingAddress: addr,
		CouponCode:      "SAVE10",
		PaymentMethod:   models.PaymentMethodUPI,
	})
	if err != nil {
		log.Printf("Place order error: %v", err)
	} else {
		fmt.Printf("Order placed: %s, Status: %s, Final: %.2f\n", order.ID, order.Status, order.FinalAmount)
	}

	// Order history
	history, _ := orderService.GetOrderHistory(ctx, user.ID, 10, 0)
	fmt.Printf("Order history: %d orders\n", len(history))
}

func seedData(ctx context.Context, productRepo *repositories.InMemoryProductRepo, categoryRepo *repositories.InMemoryCategoryRepo, userRepo *repositories.InMemoryUserRepo, couponRepo *repositories.InMemoryCouponRepo) {
	now := time.Now()

	// Category
	cat := &models.Category{
		ID: "cat-1", Name: "Electronics", Description: "Electronic devices",
		CreatedAt: now, UpdatedAt: now,
	}
	_ = categoryRepo.Create(ctx, cat)

	// Products
	prod1 := models.NewProductBuilder().
		WithID("prod-1").
		WithName("Laptop").
		WithDescription("High-performance laptop").
		WithPrice(999.99).
		WithCategoryID("cat-1").
		WithStock(10).
		WithSKU("LAP-001").
		WithRating(4.5).
		Build()
	prod1.CreatedAt = now
	prod1.UpdatedAt = now
	_ = productRepo.Create(ctx, prod1)

	prod2 := models.NewProductBuilder().
		WithID("prod-2").
		WithName("Mouse").
		WithDescription("Wireless mouse").
		WithPrice(29.99).
		WithCategoryID("cat-1").
		WithStock(100).
		WithSKU("MOU-001").
		Build()
	prod2.CreatedAt = now
	prod2.UpdatedAt = now
	_ = productRepo.Create(ctx, prod2)

	// User
	user := &models.User{
		ID: "user-1", Name: "John Doe", Email: "john@example.com", Password: "hashed",
		Phone: "1234567890", CreatedAt: now, UpdatedAt: now,
		Addresses: []models.Address{
			{ID: "addr-1", Street: "123 Main St", City: "NYC", State: "NY", Country: "USA", PostalCode: "10001", IsDefault: true},
		},
	}
	_ = userRepo.Create(ctx, user)

	// Coupon
	coupon := &models.Coupon{
		ID: "cpn-1", Code: "SAVE10", Type: models.CouponTypePercentage, Value: 10,
		MinOrderAmount: 100, ExpiresAt: now.Add(30 * 24 * time.Hour), UsageLimit: 100,
		CreatedAt: now, UpdatedAt: now,
	}
	_ = couponRepo.Create(ctx, coupon)
}
