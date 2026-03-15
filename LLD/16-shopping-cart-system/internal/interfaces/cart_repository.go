package interfaces

import "shopping-cart-system/internal/models"

// CartRepository defines data access for carts (Repository pattern)
type CartRepository interface {
	GetByUserID(userID string) (*models.Cart, error)
	Create(cart *models.Cart) error
	Update(cart *models.Cart) error
	UpdateStatus(cartID string, status models.CartStatus) error
}
