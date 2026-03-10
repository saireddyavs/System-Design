package services

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

// AuthService handles registration and authentication (SRP)
type AuthService struct {
	userRepo interfaces.UserRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo interfaces.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register creates a new user account
func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, ErrUserExists
	}
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Status:       models.StatusOffline,
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login validates credentials and returns user
func (s *AuthService) Login(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
