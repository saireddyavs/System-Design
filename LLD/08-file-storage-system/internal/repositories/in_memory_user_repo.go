package repositories

import (
	"errors"
	"file-storage-system/internal/interfaces"
	"file-storage-system/internal/models"
	"sync"
)

var ErrUserNotFound = errors.New("user not found")

// InMemoryUserRepo implements UserRepository with in-memory storage.
type InMemoryUserRepo struct {
	mu    sync.RWMutex
	users map[string]*models.User
}

// NewInMemoryUserRepo creates a new in-memory user repository.
func NewInMemoryUserRepo() interfaces.UserRepository {
	return &InMemoryUserRepo{
		users: make(map[string]*models.User),
	}
}

func (r *InMemoryUserRepo) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepo) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepo) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *InMemoryUserRepo) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}
