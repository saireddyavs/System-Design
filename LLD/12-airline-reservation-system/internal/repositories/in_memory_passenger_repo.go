package repositories

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
	"sync"
)

var (
	ErrPassengerNotFound = errors.New("passenger not found")
)

// InMemoryPassengerRepository implements PassengerRepository with in-memory storage (thread-safe)
type InMemoryPassengerRepository struct {
	passengers map[string]*models.Passenger
	mu        sync.RWMutex
}

// NewInMemoryPassengerRepository creates a new in-memory passenger repository
func NewInMemoryPassengerRepository() interfaces.PassengerRepository {
	return &InMemoryPassengerRepository{
		passengers: make(map[string]*models.Passenger),
	}
}

func (r *InMemoryPassengerRepository) Create(passenger *models.Passenger) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.passengers[passenger.ID]; exists {
		return errors.New("passenger already exists")
	}
	p := *passenger
	r.passengers[passenger.ID] = &p
	return nil
}

func (r *InMemoryPassengerRepository) GetByID(id string) (*models.Passenger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.passengers[id]
	if !exists {
		return nil, ErrPassengerNotFound
	}
	pCopy := *p
	return &pCopy, nil
}

func (r *InMemoryPassengerRepository) Update(passenger *models.Passenger) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.passengers[passenger.ID]; !exists {
		return ErrPassengerNotFound
	}
	p := *passenger
	r.passengers[passenger.ID] = &p
	return nil
}

func (r *InMemoryPassengerRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.passengers[id]; !exists {
		return ErrPassengerNotFound
	}
	delete(r.passengers, id)
	return nil
}
