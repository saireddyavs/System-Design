package tests

import (
	"atm-system/internal/hardware"
	"atm-system/internal/models"
	"context"
	"testing"
)

func TestGreedyCashDispenser_CanDispense(t *testing.T) {
	dispenser := hardware.NewGreedyCashDispenser()
	ctx := context.Background()

	inventory := models.CashInventory{
		models.Denomination100:  10,
		models.Denomination500:  10,
		models.Denomination1000: 10,
		models.Denomination2000: 10,
	}

	tests := []struct {
		amount      float64
		canDispense bool
	}{
		{100, true},
		{500, true},
		{2500, true},
		{99, false},
		{150, false},
		{36000, true},  // Total inventory: 10*100+10*500+10*1000+10*2000
		{100000, false}, // Not enough cash
	}

	for _, tt := range tests {
		got := dispenser.CanDispense(ctx, tt.amount, inventory)
		if got != tt.canDispense {
			t.Errorf("CanDispense(%v) = %v, want %v", tt.amount, got, tt.canDispense)
		}
	}
}

func TestGreedyCashDispenser_CalculateDispense_GreedyAlgorithm(t *testing.T) {
	dispenser := hardware.NewGreedyCashDispenser()
	ctx := context.Background()

	inventory := models.CashInventory{
		models.Denomination100:  10,
		models.Denomination500:  5,
		models.Denomination1000: 5,
		models.Denomination2000: 5,
	}

	// 2500 should give 1x2000 + 1x500 (largest first)
	dispensed, ok := dispenser.CalculateDispense(ctx, 2500, inventory)
	if !ok {
		t.Fatal("expected to dispense 2500")
	}
	if dispensed[models.Denomination2000] != 1 || dispensed[models.Denomination500] != 1 {
		t.Errorf("expected 1x2000 + 1x500, got %v", dispensed)
	}

	// 3000 should give 1x2000 + 1x1000
	dispensed, ok = dispenser.CalculateDispense(ctx, 3000, inventory)
	if !ok {
		t.Fatal("expected to dispense 3000")
	}
	if dispensed[models.Denomination2000] != 1 || dispensed[models.Denomination1000] != 1 {
		t.Errorf("expected 1x2000 + 1x1000, got %v", dispensed)
	}
}

func TestGreedyCashDispenser_CalculateDispense_InsufficientDenominations(t *testing.T) {
	dispenser := hardware.NewGreedyCashDispenser()
	ctx := context.Background()

	// Only 100s - cannot make 500
	inventory := models.CashInventory{
		models.Denomination100:  3,
		models.Denomination500:  0,
		models.Denomination1000: 0,
		models.Denomination2000: 0,
	}

	_, ok := dispenser.CalculateDispense(ctx, 500, inventory)
	if ok {
		t.Fatal("expected failure - cannot make 500 with only 3x100")
	}
}

func TestGreedyCashDispenser_Dispense(t *testing.T) {
	dispenser := hardware.NewGreedyCashDispenser()
	ctx := context.Background()

	inventory := models.CashInventory{
		models.Denomination100:  10,
		models.Denomination500:  10,
		models.Denomination1000: 10,
		models.Denomination2000: 10,
	}

	result, err := dispenser.Dispense(ctx, 2500, inventory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.ErrorMessage)
	}
	if result.Dispensed[models.Denomination2000] != 1 || result.Dispensed[models.Denomination500] != 1 {
		t.Errorf("expected 1x2000 + 1x500, got %v", result.Dispensed)
	}
}

func TestGreedyCashDispenser_InvalidAmount(t *testing.T) {
	dispenser := hardware.NewGreedyCashDispenser()
	ctx := context.Background()
	inventory := models.CashInventory{models.Denomination100: 10}

	result, _ := dispenser.Dispense(ctx, 150, inventory)
	if result.Success {
		t.Fatal("expected failure for non-multiple of 100")
	}
}
