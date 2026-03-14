package interfaces

import "splitwise/internal/models"

// SplitStrategy defines how to calculate splits for an expense (Strategy Pattern)
type SplitStrategy interface {
	// CalculateSplits computes the amount each participant owes
	// participants: user IDs involved in the split (excluding paidBy who gets credit)
	// amount: total expense amount
	// splitParams: type-specific params (e.g., percentages, exact amounts, shares)
	CalculateSplits(amount float64, paidBy string, participants []string, splitParams map[string]float64) ([]models.Split, error)
	// Supports returns true if this strategy handles the given split type
	Supports(splitType models.SplitType) bool
}
