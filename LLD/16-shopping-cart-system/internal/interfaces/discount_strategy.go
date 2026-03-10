package interfaces

import "shopping-cart-system/internal/models"

// DiscountContext provides data needed for discount calculation
type DiscountContext struct {
	Subtotal    float64
	Items       []models.CartItem
	Coupon      *models.Coupon
}

// DiscountStrategy defines discount calculation (Strategy pattern)
type DiscountStrategy interface {
	Calculate(ctx *DiscountContext) float64
	Supports(couponType models.CouponType) bool
}
