package services

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"social-media-platform/internal/interfaces"
	"social-media-platform/internal/models"
	"social-media-platform/internal/repositories"
)

var (
	ErrInvalidUserData = errors.New("invalid user data")
)

// UserService provides user profile operations (Facade pattern).
// Facade: Simplifies complex user profile operations behind a simple interface
type UserService struct {
	userRepo interfaces.UserRepository
	nextID   atomic.Uint64
	mu       sync.Mutex
}

// NewUserService creates a new user service
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Register creates a new user (Facade: encapsulates validation + creation)
func (s *UserService) Register(username, email, bio, profilePicURL string) (*models.User, error) {
	if username == "" || email == "" {
		return nil, ErrInvalidUserData
	}

	s.mu.Lock()
	id := fmt.Sprintf("user-%d", s.nextID.Add(1))
	s.mu.Unlock()

	user := models.NewUser(id, username, email, bio, profilePicURL)
	if err := s.userRepo.Create(user); err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return nil, fmt.Errorf("username or email already exists: %w", err)
		}
		return nil, err
	}
	return user, nil
}

// GetProfile retrieves user profile by ID
func (s *UserService) GetProfile(userID string) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}

// UpdateProfile updates user profile (Facade: encapsulates update logic)
func (s *UserService) UpdateProfile(userID, bio, profilePicURL string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("cannot update deactivated user")
	}

	user.UpdateProfile(bio, profilePicURL)
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Deactivate marks a user as inactive
func (s *UserService) Deactivate(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	user.Deactivate()
	return s.userRepo.Update(user)
}

// SearchUsers searches users by username, email, or bio
func (s *UserService) SearchUsers(query string, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.Search(query, limit)
}
