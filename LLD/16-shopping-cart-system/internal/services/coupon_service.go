package services

import (
	"fmt"
	"time"

	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
	"shopping-cart-system/internal/strategies"
)

type CouponService struct {
	repo              interfaces.CouponRepository
	discountRegistry   *strategies.DiscountStrategyRegistry
}

func NewCouponService(repo interfaces.CouponRepository, discountRegistry *strategies.DiscountStrategyRegistry) *CouponService {
	return &CouponService{
		repo:            repo,
		discountRegistry: discountRegistry,
	}
}

func (s *CouponService) ValidateAndGetDiscount(code string, subtotal float64, items []models.CartItem) (float64, *models.Coupon, error) {
	coupon, err := s.repo.GetByCode(code)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid coupon: %s", code)
	}
	if coupon.IsExpired() {
		return 0, nil, fmt.Errorf("coupon expired: %s", code)
	}
	if coupon.IsUsageExhausted() {
		return 0, nil, fmt.Errorf("coupon usage limit reached: %s", code)
	}
	if !coupon.MeetsMinOrder(subtotal) {
		return 0, nil, fmt.Errorf("minimum order amount not met: need %.2f, have %.2f", coupon.MinOrderAmount, subtotal)
	}
	strategy := s.discountRegistry.GetStrategy(coupon.Type)
	if strategy == nil {
		return 0, nil, fmt.Errorf("unsupported coupon type: %s", coupon.Type)
	}
	ctx := &interfaces.DiscountContext{
		Subtotal: subtotal,
		Items:    items,
		Coupon:   coupon,
	}
	discount := strategy.Calculate(ctx)
	return discount, coupon, nil
}

func (s *CouponService) CreateCoupon(code string, couponType models.CouponType, value, minOrder float64, expiresAt time.Time, maxUsage int) (*models.Coupon, error) {
	id := fmt.Sprintf("coupon_%d", time.Now().UnixNano())
	coupon := &models.Coupon{
		ID:             id,
		Code:           code,
		Type:           couponType,
		Value:          value,
		MinOrderAmount: minOrder,
		ExpiresAt:      expiresAt,
		MaxUsageLimit:  maxUsage,
		CurrentUsage:   0,
		CreatedAt:      time.Now(),
	}
	if err := s.repo.Create(coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

func (s *CouponService) IncrementUsage(couponID string) error {
	return s.repo.IncrementUsage(couponID)
}
