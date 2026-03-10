package interfaces

import (
	"parking-lot-system/internal/models"
	"time"
)

// FeeCalculator defines how parking fees are computed.
// Strategy pattern + ISP: Single responsibility of fee calculation.
// OCP: New fee models (flat rate, tiered, peak/off-peak) can be added
// without modifying FeeService.
type FeeCalculator interface {
	// Calculate returns the fee in cents for the given vehicle and duration.
	Calculate(vehicle models.Vehicle, duration time.Duration) int64
}
