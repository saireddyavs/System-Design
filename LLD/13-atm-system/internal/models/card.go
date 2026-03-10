package models

import "time"

// Card represents an ATM card
type Card struct {
	ID         string
	CardNumber string
	AccountID  string
	ExpiryDate time.Time
	IsBlocked  bool
	CreatedAt  time.Time
}

// NewCard creates a new card
func NewCard(id, cardNumber, accountID string, expiryDate time.Time) *Card {
	return &Card{
		ID:         id,
		CardNumber: cardNumber,
		AccountID:  accountID,
		ExpiryDate: expiryDate,
		IsBlocked:  false,
		CreatedAt:  time.Now(),
	}
}

// IsValid checks if the card is valid (not expired, not blocked)
func (c *Card) IsValid() bool {
	return !c.IsBlocked && time.Now().Before(c.ExpiryDate)
}
