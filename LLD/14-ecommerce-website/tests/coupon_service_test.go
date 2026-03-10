package tests

import (
	"context"
	"testing"
	"time"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
	"ecommerce-website/internal/services"
	"ecommerce-website/internal/strategies"
)

func setupCouponTest(t *testing.T) (context.Context, *services.CouponService) {
	ctx := context.Background()
	couponRepo := repositories.NewInMemoryCouponRepo()

	strategies := []interfaces.DiscountStrategy{
		strategies.NewPercentageDiscountStrategy(),
		strategies.NewFlatDiscountStrategy(),
		strategies.NewBOGODiscountStrategy(),
	}
	couponService := services.NewCouponService(couponRepo, strategies)

	now := time.Now()
	coupons := []*models.Coupon{
		{
			ID: "c1", Code: "SAVE10", Type: models.CouponTypePercentage, Value: 10,
			MinOrderAmount: 100, ExpiresAt: now.Add(24 * time.Hour), UsageLimit: 5,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "c2", Code: "FLAT20", Type: models.CouponTypeFlat, Value: 20,
			MinOrderAmount: 50, ExpiresAt: now.Add(24 * time.Hour), UsageLimit: 0,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "c3", Code: "BOGO", Type: models.CouponTypeBOGO, Value: 0,
			MinOrderAmount: 0, ExpiresAt: now.Add(24 * time.Hour), UsageLimit: 10,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, c := range coupons {
		_ = couponRepo.Create(ctx, c)
	}

	return ctx, couponService
}

func TestCouponService_ApplyCoupon_Percentage(t *testing.T) {
	ctx, couponService := setupCouponTest(t)

	discount, coupon, err := couponService.ApplyCoupon(ctx, "SAVE10", 200, 2)
	if err != nil {
		t.Fatalf("ApplyCoupon failed: %v", err)
	}
	if coupon == nil {
		t.Fatal("coupon is nil")
	}
	// 10% of 200 = 20
	if discount != 20 {
		t.Errorf("expected discount 20, got %.2f", discount)
	}
}

func TestCouponService_ApplyCoupon_Flat(t *testing.T) {
	ctx, couponService := setupCouponTest(t)

	discount, _, err := couponService.ApplyCoupon(ctx, "FLAT20", 100, 1)
	if err != nil {
		t.Fatalf("ApplyCoupon failed: %v", err)
	}
	if discount != 20 {
		t.Errorf("expected discount 20, got %.2f", discount)
	}
}

func TestCouponService_ApplyCoupon_BOGO(t *testing.T) {
	ctx, couponService := setupCouponTest(t)

	// 4 items at 25 each = 100. BOGO: 2 free, discount = 50
	discount, _, err := couponService.ApplyCoupon(ctx, "BOGO", 100, 4)
	if err != nil {
		t.Fatalf("ApplyCoupon failed: %v", err)
	}
	if discount != 50 {
		t.Errorf("expected discount 50 (2 free of 4), got %.2f", discount)
	}
}

func TestCouponService_Validate_MinOrderNotMet(t *testing.T) {
	ctx, couponService := setupCouponTest(t)

	_, _, err := couponService.ApplyCoupon(ctx, "SAVE10", 50, 1)
	if err != services.ErrMinOrderNotMet {
		t.Errorf("expected ErrMinOrderNotMet, got %v", err)
	}
}

func TestCouponService_Validate_NotFound(t *testing.T) {
	ctx, couponService := setupCouponTest(t)

	_, _, err := couponService.ApplyCoupon(ctx, "INVALID", 200, 1)
	if err != repositories.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
