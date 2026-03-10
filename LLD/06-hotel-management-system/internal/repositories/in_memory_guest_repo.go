package repositories

import (
	"errors"
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
	"sync"
)

var ErrGuestNotFound = errors.New("guest not found")

// InMemoryGuestRepository implements GuestRepository
type InMemoryGuestRepository struct {
	guests map[string]*models.Guest
	mu     sync.RWMutex
}

// NewInMemoryGuestRepository creates a new in-memory guest repository
func NewInMemoryGuestRepository() interfaces.GuestRepository {
	return &InMemoryGuestRepository{
		guests: make(map[string]*models.Guest),
	}
}

func (r *InMemoryGuestRepository) Create(guest *models.Guest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.guests[guest.ID]; exists {
		return errors.New("guest already exists")
	}
	r.guests[guest.ID] = guest
	return nil
}

func (r *InMemoryGuestRepository) GetByID(id string) (*models.Guest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	guest, exists := r.guests[id]
	if !exists {
		return nil, ErrGuestNotFound
	}
	return guest, nil
}

func (r *InMemoryGuestRepository) GetByEmail(email string) (*models.Guest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, g := range r.guests {
		if g.Email == email {
			return g, nil
		}
	}
	return nil, ErrGuestNotFound
}

func (r *InMemoryGuestRepository) GetAll() ([]*models.Guest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Guest, 0, len(r.guests))
	for _, g := range r.guests {
		result = append(result, g)
	}
	return result, nil
}

func (r *InMemoryGuestRepository) Update(guest *models.Guest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.guests[guest.ID]; !exists {
		return ErrGuestNotFound
	}
	r.guests[guest.ID] = guest
	return nil
}

func (r *InMemoryGuestRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.guests[id]; !exists {
		return ErrGuestNotFound
	}
	delete(r.guests, id)
	return nil
}
