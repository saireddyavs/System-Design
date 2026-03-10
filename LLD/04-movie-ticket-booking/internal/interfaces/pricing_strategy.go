package interfaces

import (
	"movie-ticket-booking/internal/models"
	"time"
)

// PricingStrategy defines price calculation (Strategy Pattern - OCP)
type PricingStrategy interface {
	CalculatePrice(basePrice float64, seatCategory models.SeatCategory, showTime time.Time) float64
}
