package interfaces

import (
	"hotel-management-system/internal/models"
	"time"
)

// PricingContext holds data needed for price calculation (Strategy pattern)
type PricingContext struct {
	Room         *models.Room
	Guest        *models.Guest
	CheckInDate  time.Time
	CheckOutDate time.Time
	Nights       int
}

// PricingStrategy defines price calculation algorithms (Strategy pattern)
// O - Open/Closed: New pricing rules (seasonal, weekend) without modifying existing code
// L - Liskov: All strategies are interchangeable
type PricingStrategy interface {
	CalculatePrice(ctx *PricingContext) float64
}
