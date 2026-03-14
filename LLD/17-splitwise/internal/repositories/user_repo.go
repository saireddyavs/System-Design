package repositories

import (
	"fmt"
	"sync"
	"time"

	"splitwise/internal/interfaces"
	"splitwise/internal/models"
)

// InMemoryUserRepository implements UserRepository with in-memory storage
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*models.User),
	}
}

// Ensure InMemoryUserRepository implements UserRepository
var _ interfaces.UserRepository = (*InMemoryUserRepository)(nil)

func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.users[user.ID] = copyUser(user)
	return nil
}

func (r *InMemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return copyUser(u), nil
}

func (r *InMemoryUserRepository) GetByIDs(ids []string) ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.User, 0, len(ids))
	for _, id := range ids {
		if u, ok := r.users[id]; ok {
			result = append(result, copyUser(u))
		}
	}
	return result, nil
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	user.UpdatedAt = time.Now()
	r.users[user.ID] = copyUser(user)
	return nil
}

func (r *InMemoryUserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	delete(r.users, id)
	return nil
}

func copyUser(u *models.User) *models.User {
	cpy := *u
	return &cpy
}
