package interfaces

import (
	"food-delivery-system/internal/models"
	"time"
)

// PricingStrategy defines the contract for order pricing (Strategy Pattern)
type PricingStrategy interface {
	CalculateDeliveryFee(restaurantLoc, deliveryAddr models.Location) float64
	CalculateSurgeFee(orderTime time.Time) float64
	CalculateTotal(subTotal, deliveryFee, surgeFee float64) float64
}
