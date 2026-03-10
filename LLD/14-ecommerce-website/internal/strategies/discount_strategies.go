package strategies

import (
	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"math"
)

// PercentageDiscountStrategy calculates percentage-based discount
type PercentageDiscountStrategy struct{}

func NewPercentageDiscountStrategy() *PercentageDiscountStrategy {
	return &PercentageDiscountStrategy{}
}

func (s *PercentageDiscountStrategy) Calculate(coupon *models.Coupon, orderAmount float64, _ int) float64 {
	if coupon.Type != models.CouponTypePercentage {
		return 0
	}
	discount := orderAmount * (coupon.Value / 100)
	return math.Round(discount*100) / 100
}

func (s *PercentageDiscountStrategy) GetType() models.CouponType {
	return models.CouponTypePercentage
}

// FlatDiscountStrategy calculates flat amount discount
type FlatDiscountStrategy struct{}

func NewFlatDiscountStrategy() *FlatDiscountStrategy {
	return &FlatDiscountStrategy{}
}

func (s *FlatDiscountStrategy) Calculate(coupon *models.Coupon, orderAmount float64, _ int) float64 {
	if coupon.Type != models.CouponTypeFlat {
		return 0
	}
	discount := math.Min(coupon.Value, orderAmount)
	return math.Round(discount*100) / 100
}

func (s *FlatDiscountStrategy) GetType() models.CouponType {
	return models.CouponTypeFlat
}

// BOGODiscountStrategy calculates buy-one-get-one discount
// For every 2 items, one gets free (50% off on pairs)
type BOGODiscountStrategy struct{}

func NewBOGODiscountStrategy() *BOGODiscountStrategy {
	return &BOGODiscountStrategy{}
}

func (s *BOGODiscountStrategy) Calculate(coupon *models.Coupon, orderAmount float64, quantity int) float64 {
	if coupon.Type != models.CouponTypeBOGO {
		return 0
	}
	// BOGO: For every 2 items, get 1 free. Discount = (quantity/2) * (orderAmount/quantity)
	if quantity < 2 {
		return 0
	}
	unitPrice := orderAmount / float64(quantity)
	freeItems := quantity / 2
	discount := float64(freeItems) * unitPrice
	return math.Round(discount*100) / 100
}

func (s *BOGODiscountStrategy) GetType() models.CouponType {
	return models.CouponTypeBOGO
}

// Ensure all strategies implement the interface
var _ interfaces.DiscountStrategy = (*PercentageDiscountStrategy)(nil)
var _ interfaces.DiscountStrategy = (*FlatDiscountStrategy)(nil)
var _ interfaces.DiscountStrategy = (*BOGODiscountStrategy)(nil)
