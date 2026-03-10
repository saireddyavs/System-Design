package interfaces

import (
	"ride-sharing-service/internal/models"
	"time"
)

// FareCalculator defines fare calculation strategy (Strategy pattern)
type FareCalculator interface {
	Calculate(pickup, dropoff models.Location, duration time.Duration, surgeMultiplier float64) float64
}
