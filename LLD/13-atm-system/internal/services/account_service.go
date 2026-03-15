package services

import (
	"atm-system/internal/interfaces"
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

func (s *AccountService) UpdatePIN(ctx context.Context, accountID string, newPINHash string) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	account.UpdatePIN(newPINHash)
	return s.accountRepo.Update(ctx, account)
}
