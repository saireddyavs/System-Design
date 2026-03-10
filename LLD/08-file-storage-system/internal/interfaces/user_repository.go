package interfaces

import "file-storage-system/internal/models"

// UserRepository defines the contract for user data access (Repository pattern).
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
}
