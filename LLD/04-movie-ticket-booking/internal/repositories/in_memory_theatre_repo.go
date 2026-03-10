package repositories

import (
	"errors"
	"movie-ticket-booking/internal/models"
	"strings"
	"sync"
)

var ErrTheatreNotFound = errors.New("theatre not found")

// InMemoryTheatreRepository implements TheatreRepository
type InMemoryTheatreRepository struct {
	theatres map[string]*models.Theatre
	mu       sync.RWMutex
}

// NewInMemoryTheatreRepository creates a new in-memory theatre repository
func NewInMemoryTheatreRepository() *InMemoryTheatreRepository {
	return &InMemoryTheatreRepository{
		theatres: make(map[string]*models.Theatre),
	}
}

func (r *InMemoryTheatreRepository) Create(theatre *models.Theatre) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.theatres[theatre.ID] = theatre
	return nil
}

func (r *InMemoryTheatreRepository) GetByID(id string) (*models.Theatre, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.theatres[id]
	if !ok {
		return nil, ErrTheatreNotFound
	}
	return t, nil
}

func (r *InMemoryTheatreRepository) GetByCity(city string) ([]*models.Theatre, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Theatre
	for _, t := range r.theatres {
		if strings.EqualFold(t.City, city) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *InMemoryTheatreRepository) GetAll() ([]*models.Theatre, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Theatre, 0, len(r.theatres))
	for _, t := range r.theatres {
		result = append(result, t)
	}
	return result, nil
}
