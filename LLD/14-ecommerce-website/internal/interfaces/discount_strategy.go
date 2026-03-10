package interfaces

import "ecommerce-website/internal/models"

// DiscountStrategy defines the interface for discount calculation (Strategy pattern)
type DiscountStrategy interface {
	Calculate(coupon *models.Coupon, orderAmount float64, quantity int) float64
	GetType() models.CouponType
}
