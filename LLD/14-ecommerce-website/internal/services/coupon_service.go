package services

import (
	"context"
	"errors"
	"time"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
)

var (
	ErrCouponExpired       = errors.New("coupon has expired")
	ErrCouponUsageExceeded = errors.New("coupon usage limit exceeded")
	ErrMinOrderNotMet      = errors.New("minimum order amount not met")
)

// CouponService handles coupon validation and discount calculation
type CouponService struct {
	repo     interfaces.CouponRepository
	strategies map[models.CouponType]interfaces.DiscountStrategy
}

// NewCouponService creates a new coupon service with discount strategies
func NewCouponService(repo interfaces.CouponRepository, strategies []interfaces.DiscountStrategy) *CouponService {
	stratMap := make(map[models.CouponType]interfaces.DiscountStrategy)
	for _, s := range strategies {
		stratMap[s.GetType()] = s
	}
	return &CouponService{
		repo:       repo,
		strategies: stratMap,
	}
}

// Validate checks if coupon is valid (expiry, min amount, usage limit)
func (s *CouponService) Validate(ctx context.Context, code string, orderAmount float64) (*models.Coupon, error) {
	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if err == repositories.ErrNotFound {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	if time.Now().After(coupon.ExpiresAt) {
		return nil, ErrCouponExpired
	}
	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		return nil, ErrCouponUsageExceeded
	}
	if orderAmount < coupon.MinOrderAmount {
		return nil, ErrMinOrderNotMet
	}

	return coupon, nil
}

// CalculateDiscount computes discount using the appropriate strategy
func (s *CouponService) CalculateDiscount(coupon *models.Coupon, orderAmount float64, quantity int) float64 {
	strategy, ok := s.strategies[coupon.Type]
	if !ok {
		return 0
	}
	return strategy.Calculate(coupon, orderAmount, quantity)
}

// ApplyCoupon validates and returns discount amount
func (s *CouponService) ApplyCoupon(ctx context.Context, code string, orderAmount float64, quantity int) (discount float64, coupon *models.Coupon, err error) {
	coupon, err = s.Validate(ctx, code, orderAmount)
	if err != nil {
		return 0, nil, err
	}
	discount = s.CalculateDiscount(coupon, orderAmount, quantity)
	return discount, coupon, nil
}

// IncrementUsage records coupon usage after successful order
func (s *CouponService) IncrementUsage(ctx context.Context, couponID string) error {
	return s.repo.IncrementUsage(ctx, couponID)
}

