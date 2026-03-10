package interfaces

// TaxContext provides data needed for tax calculation
type TaxContext struct {
	SubtotalAfterDiscount float64
	State                 string
	Country               string
}

// TaxCalculator defines tax calculation (Strategy pattern)
type TaxCalculator interface {
	Calculate(ctx *TaxContext) float64
}
