package tests

import (
	"testing"
	"time"

	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/repositories"
	"shopping-cart-system/internal/services"
	"shopping-cart-system/internal/strategies"
)

func setupCouponTest(t *testing.T) *services.CouponService {
	repo := repositories.NewInMemoryCouponRepository()
	registry := strategies.NewDiscountStrategyRegistry()
	return services.NewCouponService(repo, registry)
}

func TestCouponService_PercentageDiscount(t *testing.T) {
	svc := setupCouponTest(t)

	_, _ = svc.CreateCoupon("P10", models.CouponTypePercentage, 10, 50, time.Now().Add(24*time.Hour), 5)

	items := []models.CartItem{
		{ProductID: "p1", ProductName: "P1", UnitPrice: 100, Quantity: 1, Subtotal: 100},
	}
	discount, _, err := svc.ValidateAndGetDiscount("P10", 100, items)
	if err != nil {
		t.Fatalf("ValidateAndGetDiscount failed: %v", err)
	}
	if discount != 10 {
		t.Errorf("expected discount 10, got %.2f", discount)
	}
}

func TestCouponService_FlatDiscount(t *testing.T) {
	svc := setupCouponTest(t)

	_, _ = svc.CreateCoupon("F20", models.CouponTypeFlat, 20, 100, time.Now().Add(24*time.Hour), 5)

	items := []models.CartItem{
		{ProductID: "p1", ProductName: "P1", UnitPrice: 50, Quantity: 2, Subtotal: 100},
	}
	discount, _, err := svc.ValidateAndGetDiscount("F20", 100, items)
	if err != nil {
		t.Fatalf("ValidateAndGetDiscount failed: %v", err)
	}
	if discount != 20 {
		t.Errorf("expected discount 20, got %.2f", discount)
	}
}

func TestCouponService_BOGODiscount(t *testing.T) {
	svc := setupCouponTest(t)

	_, _ = svc.CreateCoupon("BOGO", models.CouponTypeBOGO, 1, 0, time.Now().Add(24*time.Hour), 0)

	items := []models.CartItem{
		{ProductID: "p1", ProductName: "P1", UnitPrice: 30, Quantity: 4, Subtotal: 120}, // 2 pairs -> 2 free = 60 discount
	}
	discount, _, err := svc.ValidateAndGetDiscount("BOGO", 120, items)
	if err != nil {
		t.Fatalf("ValidateAndGetDiscount failed: %v", err)
	}
	if discount != 60 {
		t.Errorf("expected BOGO discount 60 (2 free at 30 each), got %.2f", discount)
	}
}

func TestCouponService_MinOrderNotMet(t *testing.T) {
	svc := setupCouponTest(t)

	_, _ = svc.CreateCoupon("MIN100", models.CouponTypeFlat, 20, 100, time.Now().Add(24*time.Hour), 5)

	items := []models.CartItem{{ProductID: "p1", ProductName: "P1", UnitPrice: 10, Quantity: 1, Subtotal: 10}}
	_, _, err := svc.ValidateAndGetDiscount("MIN100", 50, items)
	if err == nil {
		t.Fatal("expected error for min order not met")
	}
}

func TestCouponService_InvalidCode(t *testing.T) {
	svc := setupCouponTest(t)

	_, _, err := svc.ValidateAndGetDiscount("INVALID", 100, []models.CartItem{})
	if err == nil {
		t.Fatal("expected error for invalid coupon code")
	}
}
