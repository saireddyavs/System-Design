package strategies

import (
	"errors"
	"math"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// ShareSplitStrategy splits expense by shares (e.g., 2:3:5 means 2/10, 3/10, 5/10)
type ShareSplitStrategy struct{}

// NewShareSplitStrategy creates a new ShareSplitStrategy
func NewShareSplitStrategy() *ShareSplitStrategy {
	return &ShareSplitStrategy{}
}

// Supports returns true for SHARE split type
func (s *ShareSplitStrategy) Supports(splitType models.SplitType) bool {
	return splitType == models.SplitTypeShare
}

// CalculateSplits uses shares from splitParams (key=userID, value=share as float64)
func (s *ShareSplitStrategy) CalculateSplits(amount float64, paidBy string, participants []string, splitParams map[string]float64) ([]models.Split, error) {
	if len(participants) == 0 {
		return nil, errors.New("at least one participant required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if splitParams == nil || len(splitParams) != len(participants) {
		return nil, errors.New("share must be provided for each participant")
	}

	var totalShares float64
	for _, userID := range participants {
		share, ok := splitParams[userID]
		if !ok {
			return nil, errors.New("share missing for participant: " + userID)
		}
		if share < 0 {
			return nil, errors.New("share cannot be negative")
		}
		totalShares += share
	}

	if totalShares <= 0 {
		return nil, errors.New("total shares must be positive")
	}

	splits := make([]models.Split, 0, len(participants))
	for i, userID := range participants {
		share := splitParams[userID]
		amt := math.Round(amount*(share/totalShares)*100) / 100
		// Last person gets remainder for rounding
		if i == len(participants)-1 {
			var sum float64
			for j := 0; j < len(participants)-1; j++ {
				sum += math.Round(amount*(splitParams[participants[j]]/totalShares)*100) / 100
			}
			amt = math.Round((amount-sum)*100) / 100
		}
		splits = append(splits, models.Split{
			UserID: userID,
			Amount: amt,
			Share:  int(share),
		})
	}
	return splits, nil
}

// Ensure ShareSplitStrategy implements SplitStrategy
var _ interfaces.SplitStrategy = (*ShareSplitStrategy)(nil)
