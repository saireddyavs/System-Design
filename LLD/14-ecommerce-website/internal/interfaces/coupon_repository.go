package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// CouponRepository defines the contract for coupon data access
type CouponRepository interface {
	Create(ctx context.Context, coupon *models.Coupon) error
	GetByID(ctx context.Context, id string) (*models.Coupon, error)
	GetByCode(ctx context.Context, code string) (*models.Coupon, error)
	Update(ctx context.Context, coupon *models.Coupon) error
	IncrementUsage(ctx context.Context, couponID string) error
}
