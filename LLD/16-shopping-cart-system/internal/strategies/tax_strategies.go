package strategies

import "shopping-cart-system/internal/interfaces"

// FlatTaxCalculator applies a fixed percentage tax (default 18%)
type FlatTaxCalculator struct {
	Rate float64 // e.g., 0.18 for 18%
}

func NewFlatTaxCalculator(rate float64) *FlatTaxCalculator {
	if rate <= 0 {
		rate = 0.18
	}
	return &FlatTaxCalculator{Rate: rate}
}

func (t *FlatTaxCalculator) Calculate(ctx *interfaces.TaxContext) float64 {
	return ctx.SubtotalAfterDiscount * t.Rate
}

// StateTaxCalculator applies state-specific tax rates
type StateTaxCalculator struct {
	DefaultRate float64
	StateRates  map[string]float64
}

func NewStateTaxCalculator(defaultRate float64, stateRates map[string]float64) *StateTaxCalculator {
	if defaultRate <= 0 {
		defaultRate = 0.18
	}
	return &StateTaxCalculator{
		DefaultRate: defaultRate,
		StateRates:  stateRates,
	}
}

func (t *StateTaxCalculator) Calculate(ctx *interfaces.TaxContext) float64 {
	rate := t.DefaultRate
	if r, ok := t.StateRates[ctx.State]; ok {
		rate = r
	}
	return ctx.SubtotalAfterDiscount * rate
}
