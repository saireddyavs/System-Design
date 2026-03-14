package services

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// UserService implements user management
type UserService struct {
	userRepo interfaces.UserRepository
	mu       sync.RWMutex
	idSeq    int
}

// NewUserService creates a new UserService
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		idSeq:    1,
	}
}

// Ensure UserService implements interfaces.UserService
var _ interfaces.UserService = (*UserService)(nil)

// CreateUser creates a new user with auto-generated ID
func (s *UserService) CreateUser(name, email, phone string) (*models.User, error) {
	s.mu.Lock()
	id := fmt.Sprintf("user%d", s.idSeq)
	s.idSeq++
	s.mu.Unlock()
	return s.createUserWithID(id, name, email, phone)
}

// CreateUserWithID creates a user with a specific ID (for testing/demo)
func (s *UserService) CreateUserWithID(id, name, email, phone string) (*models.User, error) {
	return s.createUserWithID(id, name, email, phone)
}

func (s *UserService) createUserWithID(id, name, email, phone string) (*models.User, error) {
	user := &models.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}
