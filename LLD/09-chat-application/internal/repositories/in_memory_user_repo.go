package repositories

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"strings"
	"sync"
)

// InMemoryUserRepository implements UserRepository with in-memory storage
type InMemoryUserRepository struct {
	users   map[string]*models.User
	byName  map[string]string // username -> id
	byEmail map[string]string // email -> id
	mu      sync.RWMutex
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() interfaces.UserRepository {
	return &InMemoryUserRepository{
		users:   make(map[string]*models.User),
		byName:  make(map[string]string),
		byEmail: make(map[string]string),
	}
}

func (r *InMemoryUserRepository) Create(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return ErrAlreadyExists
	}

	// Copy to avoid external mutation
	userCopy := *user
	r.users[user.ID] = &userCopy
	r.byName[strings.ToLower(user.Username)] = user.ID
	r.byEmail[strings.ToLower(user.Email)] = user.ID
	return nil
}

func (r *InMemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrNotFound
	}
	userCopy := *user
	return &userCopy, nil
}

func (r *InMemoryUserRepository) GetByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byName[strings.ToLower(username)]
	if !exists {
		return nil, ErrNotFound
	}
	user := r.users[id]
	userCopy := *user
	return &userCopy, nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byEmail[strings.ToLower(email)]
	if !exists {
		return nil, ErrNotFound
	}
	user := r.users[id]
	userCopy := *user
	return &userCopy, nil
}

func (r *InMemoryUserRepository) Update(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return ErrNotFound
	}

	userCopy := *user
	r.users[user.ID] = &userCopy
	return nil
}
