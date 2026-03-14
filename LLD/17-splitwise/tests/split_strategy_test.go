package tests

import (
	"testing"

	"splitwise/internal/models"
	"splitwise/internal/strategies"
)

func TestEqualSplit(t *testing.T) {
	strategy := strategies.NewEqualSplitStrategy()
	participants := []string{"u1", "u2", "u3", "u4"}

	splits, err := strategy.CalculateSplits(1000, "u1", participants, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(splits) != 4 {
		t.Fatalf("expected 4 splits, got %d", len(splits))
	}
	var total float64
	for _, s := range splits {
		total += s.Amount
	}
	if total != 1000 {
		t.Errorf("total should be 1000, got %.2f", total)
	}
	if splits[0].Amount != 250 {
		t.Errorf("each share should be 250, got %.2f", splits[0].Amount)
	}
}

func TestExactSplit(t *testing.T) {
	strategy := strategies.NewExactSplitStrategy()
	participants := []string{"u1", "u2", "u3"}
	params := map[string]float64{"u1": 100, "u2": 150, "u3": 50}

	splits, err := strategy.CalculateSplits(300, "u1", participants, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(splits) != 3 {
		t.Fatalf("expected 3 splits, got %d", len(splits))
	}
	if splits[0].Amount != 100 || splits[1].Amount != 150 || splits[2].Amount != 50 {
		t.Errorf("exact amounts wrong: got %.2f, %.2f, %.2f", splits[0].Amount, splits[1].Amount, splits[2].Amount)
	}
}

func TestExactSplitInvalidSum(t *testing.T) {
	strategy := strategies.NewExactSplitStrategy()
	participants := []string{"u1", "u2"}
	params := map[string]float64{"u1": 100, "u2": 150}

	_, err := strategy.CalculateSplits(300, "u1", participants, params)
	if err == nil {
		t.Fatal("expected error for sum != total")
	}
}

func TestPercentageSplit(t *testing.T) {
	strategy := strategies.NewPercentageSplitStrategy()
	participants := []string{"u1", "u2", "u3"}
	params := map[string]float64{"u1": 50, "u2": 30, "u3": 20}

	splits, err := strategy.CalculateSplits(1000, "u1", participants, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(splits) != 3 {
		t.Fatalf("expected 3 splits, got %d", len(splits))
	}
	var total float64
	for _, s := range splits {
		total += s.Amount
	}
	if total != 1000 {
		t.Errorf("total should be 1000, got %.2f", total)
	}
}

func TestShareSplit(t *testing.T) {
	strategy := strategies.NewShareSplitStrategy()
	participants := []string{"u1", "u2", "u3"}
	params := map[string]float64{"u1": 2, "u2": 3, "u3": 5} // 2:3:5 = 10 parts

	splits, err := strategy.CalculateSplits(1000, "u1", participants, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(splits) != 3 {
		t.Fatalf("expected 3 splits, got %d", len(splits))
	}
	var total float64
	for _, s := range splits {
		total += s.Amount
	}
	if total != 1000 {
		t.Errorf("total should be 1000, got %.2f", total)
	}
	// u1: 200, u2: 300, u3: 500
	if splits[0].Amount < 199 || splits[0].Amount > 201 {
		t.Errorf("u1 share should be ~200, got %.2f", splits[0].Amount)
	}
}

func TestRegistry(t *testing.T) {
	registry := strategies.NewSplitStrategyRegistry()

	for _, st := range []models.SplitType{models.SplitTypeEqual, models.SplitTypeExact, models.SplitTypePercentage, models.SplitTypeShare} {
		s := registry.GetStrategy(st)
		if s == nil {
			t.Errorf("no strategy for %s", st)
		}
	}
	if registry.GetStrategy(models.SplitType("INVALID")) != nil {
		t.Error("should return nil for invalid type")
	}
}
