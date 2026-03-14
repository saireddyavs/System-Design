package strategies

import (
	"errors"
	"math"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// EqualSplitStrategy splits expense equally among all participants
type EqualSplitStrategy struct{}

// NewEqualSplitStrategy creates a new EqualSplitStrategy
func NewEqualSplitStrategy() *EqualSplitStrategy {
	return &EqualSplitStrategy{}
}

// Supports returns true for EQUAL split type
func (s *EqualSplitStrategy) Supports(splitType models.SplitType) bool {
	return splitType == models.SplitTypeEqual
}

// CalculateSplits divides amount equally among participants
func (s *EqualSplitStrategy) CalculateSplits(amount float64, paidBy string, participants []string, splitParams map[string]float64) ([]models.Split, error) {
	if len(participants) == 0 {
		return nil, errors.New("at least one participant required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	perPerson := math.Round(amount/float64(len(participants))*100) / 100
	splits := make([]models.Split, 0, len(participants))

	for i, userID := range participants {
		amt := perPerson
		// Last person gets the remainder to handle rounding
		if i == len(participants)-1 {
			amt = amount - (perPerson * float64(len(participants)-1))
			if amt < 0 {
				amt = 0
			}
		}
		splits = append(splits, models.Split{
			UserID: userID,
			Amount: amt,
		})
	}
	return splits, nil
}

// Ensure EqualSplitStrategy implements SplitStrategy
var _ interfaces.SplitStrategy = (*EqualSplitStrategy)(nil)
