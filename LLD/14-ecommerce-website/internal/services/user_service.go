package services

import (
	"context"
	"time"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
	"ecommerce-website/internal/repositories"
)

// UserService handles user registration, login, profile, addresses
type UserService struct {
	repo interfaces.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo interfaces.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user
func (s *UserService) Register(ctx context.Context, user *models.User) error {
	_, err := s.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		return ErrEmailAlreadyExists
	}
	if err != repositories.ErrNotFound {
		return err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return s.repo.Create(ctx, user)
}

// GetByEmail retrieves user by email (for login)
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// GetByID retrieves user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.repo.Update(ctx, user)
}

// AddAddress adds an address to user
func (s *UserService) AddAddress(ctx context.Context, userID string, address models.Address) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if address.IsDefault {
		for i := range user.Addresses {
			user.Addresses[i].IsDefault = false
		}
	}
	user.Addresses = append(user.Addresses, address)
	user.UpdatedAt = time.Now()
	return s.repo.Update(ctx, user)
}

// SetDefaultAddress sets the default shipping address
func (s *UserService) SetDefaultAddress(ctx context.Context, userID, addressID string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	for i := range user.Addresses {
		user.Addresses[i].IsDefault = (user.Addresses[i].ID == addressID)
	}
	user.UpdatedAt = time.Now()
	return s.repo.Update(ctx, user)
}

var ErrEmailAlreadyExists = repositories.ErrAlreadyExists
