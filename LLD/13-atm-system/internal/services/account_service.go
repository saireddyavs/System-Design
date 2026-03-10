package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
)

// AccountService handles account operations
type AccountService struct {
	accountRepo interfaces.AccountRepository
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo interfaces.AccountRepository) *AccountService {
	return &AccountService{accountRepo: accountRepo}
}

func (s *AccountService) GetAccount(ctx context.Context, accountID string) (*models.Account, error) {
	return s.accountRepo.GetByID(ctx, accountID)
}

func (s *AccountService) GetBalance(ctx context.Context, accountID string) (float64, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return account.GetBalance(), nil
}

func (s *AccountService) UpdatePIN(ctx context.Context, accountID string, newPINHash string) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	account.UpdatePIN(newPINHash)
	return s.accountRepo.Update(ctx, account)
}
