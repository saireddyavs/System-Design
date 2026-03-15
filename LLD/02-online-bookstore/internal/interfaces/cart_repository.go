package interfaces

import "online-bookstore/internal/models"

// CartRepository defines the contract for cart data access.
type CartRepository interface {
	Create(cart *models.Cart) error
	GetByUserID(userID string) (*models.Cart, error)
	Update(cart *models.Cart) error
}
