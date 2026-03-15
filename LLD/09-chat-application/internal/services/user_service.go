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

