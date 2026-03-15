package hardware

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
	"errors"
	"fmt"
)

var (
	ErrInsufficientCash     = errors.New("insufficient cash in ATM")
	ErrInvalidAmount        = errors.New("amount must be multiple of 100")
	ErrCannotDispenseExact  = errors.New("cannot dispense exact amount with available denominations")
)

// GreedyCashDispenser implements cash dispensing using greedy algorithm (Strategy Pattern)
// Strategy: Always use largest denomination first to minimize note count
type GreedyCashDispenser struct{}

// NewGreedyCashDispenser creates a new greedy cash dispenser
func NewGreedyCashDispenser() *GreedyCashDispenser {
	return &GreedyCashDispenser{}
}

var _ interfaces.CashDispenser = (*GreedyCashDispenser)(nil)

// denominationOrder defines order for greedy algorithm (largest first)
var denominationOrder = []models.Denomination{
	models.Denomination2000,
	models.Denomination1000,
	models.Denomination500,
	models.Denomination100,
}

// CanDispense checks if the requested amount can be dispensed
func (d *GreedyCashDispenser) CanDispense(ctx context.Context, amount float64, inventory models.CashInventory) bool {
	if int(amount)%models.MinWithdrawalAmount != 0 {
		return false
	}
	_, ok := d.CalculateDispense(ctx, amount, inventory)
	return ok
}

// CalculateDispense calculates denomination breakdown using greedy algorithm
func (d *GreedyCashDispenser) CalculateDispense(ctx context.Context, amount float64, inventory models.CashInventory) (map[models.Denomination]int, bool) {
	if amount <= 0 || int(amount)%models.MinWithdrawalAmount != 0 {
		return nil, false
	}

	remaining := int(amount)
	result := make(map[models.Denomination]int)

	for _, denom := range denominationOrder {
		available := inventory[denom]
		if available <= 0 {
			continue
		}
		needed := remaining / int(denom)
		if needed > available {
			needed = available
		}
		if needed > 0 {
			result[denom] = needed
			remaining -= needed * int(denom)
		}
		if remaining == 0 {
			return result, true
		}
	}

	return nil, false
}

// Dispense performs the cash dispensing (modifies inventory in-place)
func (d *GreedyCashDispenser) Dispense(ctx context.Context, amount float64, inventory models.CashInventory) (*interfaces.DispenseResult, error) {
	if amount <= 0 {
		return &interfaces.DispenseResult{Success: false, ErrorMessage: ErrInvalidAmount.Error()}, nil
	}
	if int(amount)%models.MinWithdrawalAmount != 0 {
		return &interfaces.DispenseResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("amount must be multiple of %d", models.MinWithdrawalAmount),
		}, nil
	}

	dispensed, ok := d.CalculateDispense(ctx, amount, inventory)
	if !ok {
		return &interfaces.DispenseResult{
			Success:      false,
			ErrorMessage: ErrCannotDispenseExact.Error(),
		}, nil
	}

	// Update inventory (caller's responsibility in our design - we return the breakdown)
	// The ATM will update its inventory based on dispensed map
	return &interfaces.DispenseResult{
		Success:   true,
		Dispensed: dispensed,
	}, nil
}
