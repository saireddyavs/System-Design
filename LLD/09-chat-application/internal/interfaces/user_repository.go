package interfaces

import "chat-application/internal/models"

// UserRepository defines data access for users (SRP, DIP)
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Search(query string, limit int) ([]*models.User, error)
}
