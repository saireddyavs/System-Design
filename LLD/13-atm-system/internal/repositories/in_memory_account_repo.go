package repositories

import (
	"atm-system/internal/models"
	"atm-system/internal/interfaces"
	"context"
	"errors"
	"sync"
)

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrCardNotFound    = errors.New("card not found")
)

// InMemoryAccountRepository implements AccountRepository (Repository Pattern)
type InMemoryAccountRepository struct {
	accounts map[string]*models.Account
	cards    map[string]*models.Card
	accByNum map[string]string // accountNumber -> accountID
	cardToAcc map[string]string // cardNumber -> accountID
	mu       sync.RWMutex
}

// NewInMemoryAccountRepository creates a new in-memory account repository
func NewInMemoryAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{
		accounts: make(map[string]*models.Account),
		cards:    make(map[string]*models.Card),
		accByNum: make(map[string]string),
		cardToAcc: make(map[string]string),
	}
}

// Ensure InMemoryAccountRepository implements interfaces
var _ interfaces.AccountRepository = (*InMemoryAccountRepository)(nil)
var _ interfaces.CardRepository = (*InMemoryAccountRepository)(nil)

func (r *InMemoryAccountRepository) GetByID(ctx context.Context, id string) (*models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	acc, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return acc, nil
}

func (r *InMemoryAccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	accID, ok := r.accByNum[accountNumber]
	if !ok {
		return nil, ErrAccountNotFound
	}
	acc, ok := r.accounts[accID]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return acc, nil
}

func (r *InMemoryAccountRepository) Save(ctx context.Context, account *models.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accounts[account.ID] = account
	r.accByNum[account.AccountNumber] = account.ID
	return nil
}

func (r *InMemoryAccountRepository) Update(ctx context.Context, account *models.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accounts[account.ID]; !ok {
		return ErrAccountNotFound
	}
	r.accounts[account.ID] = account
	return nil
}

func (r *InMemoryAccountRepository) GetByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	card, ok := r.cards[cardNumber]
	if !ok {
		return nil, ErrCardNotFound
	}
	return card, nil
}

func (r *InMemoryAccountRepository) GetByAccountID(ctx context.Context, accountID string) (*models.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, card := range r.cards {
		if card.AccountID == accountID {
			return card, nil
		}
	}
	return nil, ErrCardNotFound
}

func (r *InMemoryAccountRepository) SaveCard(ctx context.Context, card *models.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cards[card.CardNumber] = card
	r.cardToAcc[card.CardNumber] = card.AccountID
	return nil
}

func (r *InMemoryAccountRepository) UpdateCard(ctx context.Context, card *models.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cards[card.CardNumber]; !ok {
		return ErrCardNotFound
	}
	r.cards[card.CardNumber] = card
	return nil
}
