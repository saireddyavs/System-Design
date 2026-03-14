package strategies

import (
	"errors"
	"math"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// PercentageSplitStrategy splits expense by percentage per participant
type PercentageSplitStrategy struct{}

// NewPercentageSplitStrategy creates a new PercentageSplitStrategy
func NewPercentageSplitStrategy() *PercentageSplitStrategy {
	return &PercentageSplitStrategy{}
}

// Supports returns true for PERCENTAGE split type
func (s *PercentageSplitStrategy) Supports(splitType models.SplitType) bool {
	return splitType == models.SplitTypePercentage
}

// CalculateSplits uses percentages from splitParams (key=userID, value=percentage 0-100)
func (s *PercentageSplitStrategy) CalculateSplits(amount float64, paidBy string, participants []string, splitParams map[string]float64) ([]models.Split, error) {
	if len(participants) == 0 {
		return nil, errors.New("at least one participant required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if splitParams == nil || len(splitParams) != len(participants) {
		return nil, errors.New("percentage must be provided for each participant")
	}

	var totalPct float64
	splits := make([]models.Split, 0, len(participants))

	for _, userID := range participants {
		pct, ok := splitParams[userID]
		if !ok {
			return nil, errors.New("percentage missing for participant: " + userID)
		}
		if pct < 0 || pct > 100 {
			return nil, errors.New("percentage must be between 0 and 100")
		}
		totalPct += pct
		amt := math.Round(amount*(pct/100)*100) / 100
		splits = append(splits, models.Split{
			UserID:     userID,
			Amount:     amt,
			Percentage: pct,
		})
	}

	// Percentages must sum to 100 (allow small tolerance)
	if math.Abs(totalPct-100) > 0.01 {
		return nil, errors.New("percentages must sum to 100")
	}

	// Adjust last split for rounding
	if len(splits) > 0 {
		var sum float64
		for i := 0; i < len(splits)-1; i++ {
			sum += splits[i].Amount
		}
		splits[len(splits)-1].Amount = math.Round((amount-sum)*100) / 100
	}

	return splits, nil
}

// Ensure PercentageSplitStrategy implements SplitStrategy
var _ interfaces.SplitStrategy = (*PercentageSplitStrategy)(nil)
