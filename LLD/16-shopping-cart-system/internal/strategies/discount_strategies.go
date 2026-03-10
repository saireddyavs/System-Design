package strategies

import (
	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
)

// PercentageDiscountStrategy applies percentage off (e.g., 10% off)
type PercentageDiscountStrategy struct{}

func (s *PercentageDiscountStrategy) Supports(couponType models.CouponType) bool {
	return couponType == models.CouponTypePercentage
}

func (s *PercentageDiscountStrategy) Calculate(ctx *interfaces.DiscountContext) float64 {
	if ctx.Coupon == nil || ctx.Coupon.Type != models.CouponTypePercentage {
		return 0
	}
	// Value is 0-100 percentage
	discount := ctx.Subtotal * (ctx.Coupon.Value / 100)
	if discount > ctx.Subtotal {
		return ctx.Subtotal
	}
	return discount
}

// FlatDiscountStrategy applies fixed amount off (e.g., $10 off)
type FlatDiscountStrategy struct{}

func (s *FlatDiscountStrategy) Supports(couponType models.CouponType) bool {
	return couponType == models.CouponTypeFlat
}

func (s *FlatDiscountStrategy) Calculate(ctx *interfaces.DiscountContext) float64 {
	if ctx.Coupon == nil || ctx.Coupon.Type != models.CouponTypeFlat {
		return 0
	}
	discount := ctx.Coupon.Value
	if discount > ctx.Subtotal {
		return ctx.Subtotal
	}
	return discount
}

// BOGODiscountStrategy applies Buy One Get One (free cheapest item)
type BOGODiscountStrategy struct{}

func (s *BOGODiscountStrategy) Supports(couponType models.CouponType) bool {
	return couponType == models.CouponTypeBOGO
}

func (s *BOGODiscountStrategy) Calculate(ctx *interfaces.DiscountContext) float64 {
	if ctx.Coupon == nil || ctx.Coupon.Type != models.CouponTypeBOGO || len(ctx.Items) == 0 {
		return 0
	}
	// BOGO: For every 2 items of same product (or Value specifies ratio), give discount
	// Simplified: Value = number of free items per pair (e.g., 1 = buy 1 get 1 free)
	// For mixed cart: apply to cheapest item per pair
	var totalDiscount float64
	for _, item := range ctx.Items {
		pairs := item.Quantity / 2
		if pairs > 0 {
			// Free items = pairs (BOGO = 1 free per 2 bought)
			freeQty := pairs
			totalDiscount += item.UnitPrice * float64(freeQty)
		}
	}
	if totalDiscount > ctx.Subtotal {
		return ctx.Subtotal
	}
	return totalDiscount
}

// DiscountStrategyRegistry holds all strategies and selects by coupon type
type DiscountStrategyRegistry struct {
	strategies []interfaces.DiscountStrategy
}

func NewDiscountStrategyRegistry() *DiscountStrategyRegistry {
	return &DiscountStrategyRegistry{
		strategies: []interfaces.DiscountStrategy{
			&PercentageDiscountStrategy{},
			&FlatDiscountStrategy{},
			&BOGODiscountStrategy{},
		},
	}
}

func (r *DiscountStrategyRegistry) GetStrategy(couponType models.CouponType) interfaces.DiscountStrategy {
	for _, s := range r.strategies {
		if s.Supports(couponType) {
			return s
		}
	}
	return nil
}
