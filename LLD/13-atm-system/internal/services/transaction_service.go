package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"

	"github.com/google/uuid"
)

// TransactionService handles transaction operations
type TransactionService struct {
	txRepo interfaces.TransactionRepository
}

// NewTransactionService creates a new transaction service
func NewTransactionService(txRepo interfaces.TransactionRepository) *TransactionService {
	return &TransactionService{txRepo: txRepo}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, accountID string, txType models.TransactionType, amount float64, balanceBefore, balanceAfter float64) (*models.Transaction, error) {
	tx := models.NewTransaction(uuid.New().String(), accountID, txType, amount, balanceBefore, balanceAfter)
	tx.Status = models.TransactionStatusCompleted
	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *TransactionService) GetMiniStatement(ctx context.Context, accountID string, limit int) ([]*models.Transaction, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.txRepo.GetByAccountID(ctx, accountID, limit)
}
