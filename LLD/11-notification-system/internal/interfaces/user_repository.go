package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// UserRepository defines operations for user data
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	Save(ctx context.Context, user *models.User) error
}
