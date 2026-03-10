package repositories

import (
	"errors"
	"movie-ticket-booking/internal/models"
	"sync"
)

var ErrScreenNotFound = errors.New("screen not found")

// InMemoryScreenRepository implements ScreenRepository
type InMemoryScreenRepository struct {
	screens map[string]*models.Screen
	mu      sync.RWMutex
}

// NewInMemoryScreenRepository creates a new in-memory screen repository
func NewInMemoryScreenRepository() *InMemoryScreenRepository {
	return &InMemoryScreenRepository{
		screens: make(map[string]*models.Screen),
	}
}

func (r *InMemoryScreenRepository) Create(screen *models.Screen) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.screens[screen.ID] = screen
	return nil
}

func (r *InMemoryScreenRepository) GetByID(id string) (*models.Screen, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.screens[id]
	if !ok {
		return nil, ErrScreenNotFound
	}
	return s, nil
}

func (r *InMemoryScreenRepository) GetByTheatreID(theatreID string) ([]*models.Screen, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Screen
	for _, s := range r.screens {
		if s.TheatreID == theatreID {
			result = append(result, s)
		}
	}
	return result, nil
}
