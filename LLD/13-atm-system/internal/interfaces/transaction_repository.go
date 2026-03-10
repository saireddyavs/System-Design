package interfaces

import (
	"atm-system/internal/models"
	"context"
	"time"
)

// TransactionRepository defines the interface for transaction data access (Repository Pattern)
// SRP: Single responsibility - only transaction persistence
type TransactionRepository interface {
	Save(ctx context.Context, tx *models.Transaction) error
	GetByAccountID(ctx context.Context, accountID string, limit int) ([]*models.Transaction, error)
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	GetByAccountIDAndDateRange(ctx context.Context, accountID string, from, to time.Time) ([]*models.Transaction, error)
}
