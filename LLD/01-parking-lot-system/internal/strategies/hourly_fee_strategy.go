package strategies

import (
	"parking-lot-system/internal/models"
	"time"
)

// HourlyFeeStrategy calculates fees based on hourly rates.
// Different rates per vehicle type. Minimum 1 hour charge.
// Rates in cents: Motorcycle=10/hr, Car=20/hr, Bus=50/hr.
type HourlyFeeStrategy struct {
	ratesPerHour map[models.VehicleType]int64
}

// NewHourlyFeeStrategy creates strategy with default rates (cents per hour).
func NewHourlyFeeStrategy() *HourlyFeeStrategy {
	return &HourlyFeeStrategy{
		ratesPerHour: map[models.VehicleType]int64{
			models.VehicleTypeMotorcycle: 1000,  // $10/hr
			models.VehicleTypeCar:         2000,  // $20/hr
			models.VehicleTypeBus:         5000,  // $50/hr
		},
	}
}

// NewHourlyFeeStrategyWithRates allows custom rates for testing/flexibility.
func NewHourlyFeeStrategyWithRates(rates map[models.VehicleType]int64) *HourlyFeeStrategy {
	return &HourlyFeeStrategy{ratesPerHour: rates}
}

// Calculate returns fee in cents. Rounds up partial hours.
func (h *HourlyFeeStrategy) Calculate(vehicle models.Vehicle, duration time.Duration) int64 {
	rate, ok := h.ratesPerHour[vehicle.GetType()]
	if !ok {
		rate = h.ratesPerHour[models.VehicleTypeCar]
	}
	hours := duration.Hours()
	if hours < 1 {
		hours = 1
	}
	// Ceiling for partial hours
	wholeHours := int64(hours)
	if hours > float64(wholeHours) {
		wholeHours++
	}
	return rate * wholeHours
}
