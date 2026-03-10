package interfaces

import (
	"library-management-system/internal/models"
	"time"
)

// FineCalculator defines fine calculation strategy (Strategy pattern)
// Different strategies: per-day, flat rate, tiered, etc.
type FineCalculator interface {
	Calculate(loan *models.Loan, asOf time.Time) float64
}
