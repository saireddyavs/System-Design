package services

import (
	"parking-lot-system/internal/interfaces"
	"parking-lot-system/internal/models"
	"time"
)

// FeeService calculates parking fees using the configured strategy.
// DIP: Depends on FeeCalculator interface, not concrete implementation.
// SRP: Only responsible for fee calculation.
type FeeService struct {
	calculator interfaces.FeeCalculator
}

// NewFeeService creates a fee service with the given calculator.
func NewFeeService(calculator interfaces.FeeCalculator) *FeeService {
	return &FeeService{calculator: calculator}
}

// CalculateFee returns the fee in cents for the given ticket and duration.
// If duration is zero, uses ticket's entry time to now.
func (f *FeeService) CalculateFee(ticket *models.Ticket, duration time.Duration) int64 {
	if duration == 0 {
		duration = ticket.GetDuration()
	}
	return f.calculator.Calculate(ticket.Vehicle, duration)
}

// CalculateFeeForVehicle returns the fee for a vehicle and duration directly.
// Useful when unparking without a ticket (e.g., lost ticket scenario).
func (f *FeeService) CalculateFeeForVehicle(vehicle models.Vehicle, duration time.Duration) int64 {
	return f.calculator.Calculate(vehicle, duration)
}
