package services

import (
	"library-management-system/internal/models"
	"time"
)

// PerDayFineCalculator implements $1 per day overdue (Strategy pattern)
type PerDayFineCalculator struct {
	RatePerDay float64
}

// NewPerDayFineCalculator creates calculator with default $1/day rate
func NewPerDayFineCalculator() *PerDayFineCalculator {
	return &PerDayFineCalculator{RatePerDay: 1.0}
}

// Calculate returns fine amount based on days overdue
func (c *PerDayFineCalculator) Calculate(loan *models.Loan, asOf time.Time) float64 {
	if loan.ReturnDate != nil {
		// Use return date if already returned
		asOf = *loan.ReturnDate
	}
	if asOf.Before(loan.DueDate) || asOf.Equal(loan.DueDate) {
		return 0
	}
	days := int(asOf.Sub(loan.DueDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return float64(days) * c.RatePerDay
}
