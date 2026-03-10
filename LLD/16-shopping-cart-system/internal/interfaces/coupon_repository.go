package interfaces

import "shopping-cart-system/internal/models"

// CouponRepository defines data access for coupons
type CouponRepository interface {
	GetByCode(code string) (*models.Coupon, error)
	GetByID(id string) (*models.Coupon, error)
	Create(coupon *models.Coupon) error
	Update(coupon *models.Coupon) error
	IncrementUsage(couponID string) error
}
