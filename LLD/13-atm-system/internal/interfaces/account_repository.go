package interfaces

import (
	"atm-system/internal/models"
	"context"
)

// AccountRepository defines the interface for account data access (Repository Pattern)
// SRP: Single responsibility - only account persistence
// DIP: High-level modules depend on abstraction, not concrete implementation
type AccountRepository interface {
	GetByID(ctx context.Context, id string) (*models.Account, error)
	Save(ctx context.Context, account *models.Account) error
	Update(ctx context.Context, account *models.Account) error
}

// CardRepository defines the interface for card data access
type CardRepository interface {
	GetByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error)
	GetByAccountID(ctx context.Context, accountID string) (*models.Card, error)
	SaveCard(ctx context.Context, card *models.Card) error
	UpdateCard(ctx context.Context, card *models.Card) error
}
