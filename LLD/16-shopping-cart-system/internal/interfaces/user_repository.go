package interfaces

import "shopping-cart-system/internal/models"

// UserRepository defines data access for users
type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
}
