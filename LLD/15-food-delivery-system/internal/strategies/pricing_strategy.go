package strategies

import (
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
	"math"
	"time"
)

const (
	BaseDeliveryFeePerKm = 5.0
	MinDeliveryFee       = 20.0
	MaxDeliveryFee       = 100.0
	PeakHourSurgePercent = 0.2 // 20% surge during peak hours
)

// DefaultPricingStrategy implements base + delivery fee + surge pricing
type DefaultPricingStrategy struct{}

// NewDefaultPricingStrategy creates a new pricing strategy
func NewDefaultPricingStrategy() interfaces.PricingStrategy {
	return &DefaultPricingStrategy{}
}

// CalculateDeliveryFee computes fee based on distance (restaurant to delivery address)
func (s *DefaultPricingStrategy) CalculateDeliveryFee(restaurantLoc, deliveryAddr models.Location) float64 {
	distanceKm := restaurantLoc.Distance(deliveryAddr)
	fee := BaseDeliveryFeePerKm * distanceKm
	fee = math.Max(fee, MinDeliveryFee)
	fee = math.Min(fee, MaxDeliveryFee)
	return math.Round(fee*100) / 100
}

// CalculateSurgeFee returns surge during peak hours (12-2 PM, 7-9 PM)
func (s *DefaultPricingStrategy) CalculateSurgeFee(orderTime time.Time) float64 {
	hour := orderTime.Hour()
	isPeak := (hour >= 12 && hour < 14) || (hour >= 19 && hour < 21)
	if isPeak {
		return PeakHourSurgePercent
	}
	return 0
}

// CalculateTotal computes final total (subTotal + deliveryFee + surgeFee * subTotal)
func (s *DefaultPricingStrategy) CalculateTotal(subTotal, deliveryFee, surgeFee float64) float64 {
	surgeAmount := subTotal * surgeFee
	total := subTotal + deliveryFee + surgeAmount
	return math.Round(total*100) / 100
}
