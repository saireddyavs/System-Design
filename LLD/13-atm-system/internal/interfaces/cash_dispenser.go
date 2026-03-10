package interfaces

import (
	"atm-system/internal/models"
	"context"
)

// DispenseResult contains the result of a cash dispense operation
type DispenseResult struct {
	Success      bool
	Dispensed    map[models.Denomination]int
	ErrorMessage string
}

// CashDispenser defines the interface for cash dispensing (Strategy Pattern - interface for algorithms)
// OCP: Open for extension - can add new dispensing strategies without modifying existing code
// ISP: Interface segregation - only dispensing-related methods
type CashDispenser interface {
	// CanDispense checks if the requested amount can be dispensed with available denominations
	CanDispense(ctx context.Context, amount float64, inventory models.CashInventory) bool

	// CalculateDispense calculates the denomination breakdown for an amount (greedy algorithm)
	CalculateDispense(ctx context.Context, amount float64, inventory models.CashInventory) (map[models.Denomination]int, bool)

	// Dispense performs the actual dispensing (updates inventory)
	Dispense(ctx context.Context, amount float64, inventory models.CashInventory) (*DispenseResult, error)
}
