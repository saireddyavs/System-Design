package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

var (
	ErrUserExists  = errors.New("user with email already exists")
	ErrInvalidAuth = errors.New("invalid email or password")
)

// AuthService handles user registration and authentication.
// SRP: Authentication concerns only.
type AuthService struct {
	userRepo interfaces.UserRepository
}

func NewAuthService(userRepo interfaces.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(name, email, password, address string) (*models.User, error) {
	existing, _ := s.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, ErrUserExists
	}

	user := &models.User{
		ID:        generateUserID(),
		Name:      name,
		Email:     email,
		Password:  hashPassword(password), // In production: use bcrypt
		Address:   address,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil || user == nil {
		return nil, ErrInvalidAuth
	}
	if !verifyPassword(password, user.Password) {
		return nil, ErrInvalidAuth
	}
	return user, nil
}

// Simplified auth - in production use bcrypt
func hashPassword(password string) string {
	return password // Placeholder - use bcrypt in production
}

func verifyPassword(plain, hashed string) bool {
	return plain == hashed
}

func generateUserID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
