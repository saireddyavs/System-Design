package interfaces

import (
	"atm-system/internal/models"
	"context"
)

// CommandResult contains the result of executing an ATM command
type CommandResult struct {
	Success     bool
	Message     string
	Transaction *models.Transaction
	Data        interface{} // For balance, mini statement, etc.
}

// ATMCommand defines the interface for ATM operations (Command Pattern)
// Encapsulates each operation as an object - enables undo, logging, queueing
type ATMCommand interface {
	Execute(ctx context.Context) (*CommandResult, error)
	GetType() models.TransactionType
}
