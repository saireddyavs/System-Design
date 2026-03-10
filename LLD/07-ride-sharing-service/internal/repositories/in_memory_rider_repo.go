package repositories

import (
	"errors"
	"ride-sharing-service/internal/models"
	"sync"
)

var ErrRiderNotFound = errors.New("rider not found")

// InMemoryRiderRepository implements RiderRepository with thread-safe in-memory storage
type InMemoryRiderRepository struct {
	riders map[string]*models.Rider
	mu     sync.RWMutex
}

// NewInMemoryRiderRepository creates a new in-memory rider repository
func NewInMemoryRiderRepository() *InMemoryRiderRepository {
	return &InMemoryRiderRepository{
		riders: make(map[string]*models.Rider),
	}
}

// Create adds a new rider
func (r *InMemoryRiderRepository) Create(rider *models.Rider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.riders[rider.ID] = rider
	return nil
}

// GetByID retrieves a rider by ID
func (r *InMemoryRiderRepository) GetByID(id string) (*models.Rider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rider, ok := r.riders[id]
	if !ok {
		return nil, ErrRiderNotFound
	}
	return rider, nil
}

// Update updates an existing rider
func (r *InMemoryRiderRepository) Update(rider *models.Rider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.riders[rider.ID]; !ok {
		return ErrRiderNotFound
	}
	r.riders[rider.ID] = rider
	return nil
}
