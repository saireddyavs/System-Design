package services

import (
	"context"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// UserService handles user lookup for the demo flow
type UserService struct {
	repo interfaces.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo interfaces.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetByEmail retrieves user by email (for login)
func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}
