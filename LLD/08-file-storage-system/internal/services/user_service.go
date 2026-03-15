package services

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
)

var ErrUserAlreadyExists = errors.New("user already exists")

// UserService handles user management and storage quota.
type UserService struct {
	mu       sync.RWMutex
	userRepo interfaces.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with default quota.
func (s *UserService) CreateUser(id, name, email string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, _ := s.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}
	user := models.NewUser(id, name, email)
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userRepo.GetByID(id)
}

// CheckQuota verifies if user has enough storage for the given size.
func (s *UserService) CheckQuota(userID string, additionalSize int64) (bool, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}
	return user.GetAvailableStorage() >= additionalSize, nil
}
