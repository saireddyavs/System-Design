package services

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserService handles user operations (SRP)
type UserService struct {
	userRepo interfaces.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUser retrieves user by ID
func (s *UserService) GetUser(userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// SetOnline updates user status to online
func (s *UserService) SetOnline(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	user.Status = models.StatusOnline
	user.LastSeen = time.Now()
	return s.userRepo.Update(user)
}

// SetOffline updates user status to offline
func (s *UserService) SetOffline(userID string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	user.Status = models.StatusOffline
	user.LastSeen = time.Now()
	return s.userRepo.Update(user)
}

// SearchUsers searches users by username or email
func (s *UserService) SearchUsers(query string, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.userRepo.Search(query, limit)
}
