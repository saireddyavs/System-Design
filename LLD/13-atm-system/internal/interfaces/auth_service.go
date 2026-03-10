package interfaces

import (
	"atm-system/internal/models"
	"context"
)

// AuthResult contains authentication result
type AuthResult struct {
	Success   bool
	Account   *models.Account
	Card      *models.Card
	Message   string
	AttemptsLeft int
}

// AuthService defines the interface for authentication (Strategy/Service abstraction)
// DIP: Services depend on this abstraction
type AuthService interface {
	// Authenticate validates card number and PIN
	Authenticate(ctx context.Context, cardNumber, pin string) (*AuthResult, error)

	// ValidateCard checks if card is valid (not blocked, not expired)
	ValidateCard(ctx context.Context, cardNumber string) (*models.Card, error)
}
