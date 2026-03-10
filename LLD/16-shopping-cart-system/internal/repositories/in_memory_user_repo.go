package repositories

import (
	"fmt"
	"sync"

	"shopping-cart-system/internal/models"
)

type InMemoryUserRepository struct {
	mu       sync.RWMutex
	users    map[string]*models.User
	byEmail  map[string]string
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:   make(map[string]*models.User),
		byEmail: make(map[string]string),
	}
}

func (r *InMemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	cpy := *u
	return &cpy, nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", email)
	}
	return r.GetByID(id)
}

func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}
	cpy := *user
	r.users[user.ID] = &cpy
	r.byEmail[user.Email] = user.ID
	return nil
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.users[user.ID]
	if !ok {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	if old.Email != user.Email {
		delete(r.byEmail, old.Email)
	}
	cpy := *user
	r.users[user.ID] = &cpy
	r.byEmail[user.Email] = user.ID
	return nil
}
