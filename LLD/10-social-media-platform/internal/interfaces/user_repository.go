package interfaces

import (
	"social-media-platform/internal/models"
)

// UserRepository defines the contract for user data access (Repository pattern).
// D - Dependency Inversion: Services depend on this interface, not concrete implementation
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Search(query string, limit int) ([]*models.User, error)
}

// FriendshipRepository defines the contract for friendship data access.
type FriendshipRepository interface {
	Create(friendship *models.Friendship) error
	GetByID(id string) (*models.Friendship, error)
	GetPendingForUser(userID string) ([]*models.Friendship, error)
	GetAcceptedFriends(userID string) ([]string, error)
	GetFollowers(userID string) ([]string, error)
	GetFollowing(userID string) ([]string, error)
	Update(friendship *models.Friendship) error
	Delete(requesterID, receiverID string) error
	GetFriendship(requesterID, receiverID string) (*models.Friendship, error)
}
