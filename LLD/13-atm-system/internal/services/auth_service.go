package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
	"errors"
	"sync"
)

var (
	ErrCardBlocked     = errors.New("card is blocked")
	ErrCardExpired     = errors.New("card is expired")
	ErrInvalidCredentials = errors.New("invalid card number or PIN")
)

// AuthService implements user authentication
type AuthService struct {
	accountRepo interfaces.AccountRepository
	cardRepo    interfaces.CardRepository
	pinAttempts map[string]int // cardNumber -> wrong attempts
	mu          sync.RWMutex
}

// NewAuthService creates a new auth service
func NewAuthService(accountRepo interfaces.AccountRepository, cardRepo interfaces.CardRepository) *AuthService {
	return &AuthService{
		accountRepo: accountRepo,
		cardRepo:    cardRepo,
		pinAttempts: make(map[string]int),
	}
}

var _ interfaces.AuthService = (*AuthService)(nil)

func (s *AuthService) ValidateCard(ctx context.Context, cardNumber string) (*models.Card, error) {
	card, err := s.cardRepo.GetByCardNumber(ctx, cardNumber)
	if err != nil {
		return nil, err
	}
	if card.IsBlocked {
		return nil, ErrCardBlocked
	}
	if !card.IsValid() {
		return nil, ErrCardExpired
	}
	return card, nil
}

func (s *AuthService) Authenticate(ctx context.Context, cardNumber, pin string) (*interfaces.AuthResult, error) {
	card, err := s.ValidateCard(ctx, cardNumber)
	if err != nil {
		return &interfaces.AuthResult{Success: false, Message: err.Error()}, nil
	}

	account, err := s.accountRepo.GetByID(ctx, card.AccountID)
	if err != nil {
		return &interfaces.AuthResult{Success: false, Message: "account not found"}, nil
	}

	s.mu.Lock()
	attempts := s.pinAttempts[cardNumber]
	s.mu.Unlock()

	if attempts >= models.MaxPINAttempts {
		card.IsBlocked = true
		_ = s.cardRepo.UpdateCard(ctx, card)
		return &interfaces.AuthResult{
			Success:     false,
			Message:     "card blocked due to too many wrong PIN attempts",
			AttemptsLeft: 0,
		}, nil
	}

	if !account.VerifyPIN(pin) {
		s.mu.Lock()
		s.pinAttempts[cardNumber] = attempts + 1
		remaining := models.MaxPINAttempts - (attempts + 1)
		s.mu.Unlock()

		return &interfaces.AuthResult{
			Success:     false,
			Message:     "invalid PIN",
			AttemptsLeft: remaining,
		}, nil
	}

	// Success - reset attempts
	s.mu.Lock()
	delete(s.pinAttempts, cardNumber)
	s.mu.Unlock()

	return &interfaces.AuthResult{
		Success:     true,
		Account:     account,
		Card:        card,
		AttemptsLeft: models.MaxPINAttempts,
	}, nil
}
