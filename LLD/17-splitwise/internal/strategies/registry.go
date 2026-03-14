package strategies

import (
	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// SplitStrategyRegistry holds all split strategies (Factory-like selection)
type SplitStrategyRegistry struct {
	strategies []interfaces.SplitStrategy
}

// NewSplitStrategyRegistry creates a registry with all built-in strategies
func NewSplitStrategyRegistry() *SplitStrategyRegistry {
	return &SplitStrategyRegistry{
		strategies: []interfaces.SplitStrategy{
			NewEqualSplitStrategy(),
			NewExactSplitStrategy(),
			NewPercentageSplitStrategy(),
			NewShareSplitStrategy(),
		},
	}
}

// GetStrategy returns the strategy for the given split type
func (r *SplitStrategyRegistry) GetStrategy(splitType models.SplitType) interfaces.SplitStrategy {
	for _, s := range r.strategies {
		if s.Supports(splitType) {
			return s
		}
	}
	return nil
}
