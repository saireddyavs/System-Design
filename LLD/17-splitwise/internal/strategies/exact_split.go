package strategies

import (
	"errors"
	"fmt"
	"math"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// ExactSplitStrategy splits expense by exact amounts per participant
type ExactSplitStrategy struct{}

// NewExactSplitStrategy creates a new ExactSplitStrategy
func NewExactSplitStrategy() *ExactSplitStrategy {
	return &ExactSplitStrategy{}
}

// Supports returns true for EXACT split type
func (s *ExactSplitStrategy) Supports(splitType models.SplitType) bool {
	return splitType == models.SplitTypeExact
}

// CalculateSplits uses exact amounts from splitParams (key=userID, value=amount)
func (s *ExactSplitStrategy) CalculateSplits(amount float64, paidBy string, participants []string, splitParams map[string]float64) ([]models.Split, error) {
	if len(participants) == 0 {
		return nil, errors.New("at least one participant required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if splitParams == nil || len(splitParams) != len(participants) {
		return nil, errors.New("exact amounts must be provided for each participant")
	}

	var sum float64
	splits := make([]models.Split, 0, len(participants))

	for _, userID := range participants {
		amt, ok := splitParams[userID]
		if !ok {
			return nil, errors.New("exact amount missing for participant: " + userID)
		}
		if amt < 0 {
			return nil, errors.New("exact amount cannot be negative")
		}
		sum += amt
		splits = append(splits, models.Split{
			UserID:      userID,
			Amount:      math.Round(amt*100) / 100,
			ExactAmount: amt,
		})
	}

	// Allow small floating point tolerance (0.01)
	if math.Abs(sum-amount) > 0.01 {
		return nil, fmt.Errorf("exact amounts must sum to total: got %.2f, expected %.2f", sum, amount)
	}

	return splits, nil
}

// Ensure ExactSplitStrategy implements SplitStrategy
var _ interfaces.SplitStrategy = (*ExactSplitStrategy)(nil)
