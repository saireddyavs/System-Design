package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryUserRepo implements UserRepository with thread-safe in-memory storage
type InMemoryUserRepo struct {
	mu     sync.RWMutex
	users  map[string]*models.User
	byEmail map[string]string
}

// NewInMemoryUserRepo creates a new in-memory user repository
func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		users:   make(map[string]*models.User),
		byEmail: make(map[string]string),
	}
}

var _ interfaces.UserRepository = (*InMemoryUserRepo)(nil)

func (r *InMemoryUserRepo) Create(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrAlreadyExists
	}
	if _, exists := r.byEmail[user.Email]; exists {
		return ErrAlreadyExists
	}
	r.users[user.ID] = user
	r.byEmail[user.Email] = user.ID
	return nil
}

func (r *InMemoryUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyUser(u), nil
}

func (r *InMemoryUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	return copyUser(r.users[id]), nil
}

func (r *InMemoryUserRepo) Update(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[user.ID]
	if !ok {
		return ErrNotFound
	}
	if existing.Email != user.Email {
		delete(r.byEmail, existing.Email)
		r.byEmail[user.Email] = user.ID
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return ErrNotFound
	}
	delete(r.byEmail, u.Email)
	delete(r.users, id)
	return nil
}

func copyUser(u *models.User) *models.User {
	cp := *u
	cp.Addresses = make([]models.Address, len(u.Addresses))
	copy(cp.Addresses, u.Addresses)
	return &cp
}
