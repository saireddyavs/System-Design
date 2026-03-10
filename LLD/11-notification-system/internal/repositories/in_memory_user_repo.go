package repositories

import (
	"context"
	"sync"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// InMemoryUserRepository is a thread-safe in-memory user store
type InMemoryUserRepository struct {
	users map[string]*models.User
	mu    sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repo
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*models.User),
	}
}

// GetByID retrieves a user by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// Save stores or updates a user
func (r *InMemoryUserRepository) Save(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

var _ interfaces.UserRepository = (*InMemoryUserRepository)(nil)
