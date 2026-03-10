package interfaces

import (
	"atm-system/internal/models"
	"context"
)

// ValidationResult contains the result of transaction validation (Chain of Responsibility)
type ValidationResult struct {
	Valid   bool
	Message string
}

// TransactionValidator defines the interface for validation chain (Chain of Responsibility Pattern)
// Each validator can pass to the next in the chain
type TransactionValidator interface {
	// Validate checks the transaction and optionally passes to next validator
	Validate(ctx context.Context, account *models.Account, amount float64, txType models.TransactionType) *ValidationResult
	SetNext(validator TransactionValidator)
}
