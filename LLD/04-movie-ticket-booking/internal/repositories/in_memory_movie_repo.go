package repositories

import (
	"errors"
	"movie-ticket-booking/internal/models"
	"strings"
	"sync"
)

var ErrMovieNotFound = errors.New("movie not found")

// InMemoryMovieRepository implements MovieRepository (thread-safe)
type InMemoryMovieRepository struct {
	movies map[string]*models.Movie
	mu     sync.RWMutex
}

// NewInMemoryMovieRepository creates a new in-memory movie repository
func NewInMemoryMovieRepository() *InMemoryMovieRepository {
	return &InMemoryMovieRepository{
		movies: make(map[string]*models.Movie),
	}
}

func (r *InMemoryMovieRepository) Create(movie *models.Movie) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.movies[movie.ID] = movie
	return nil
}

func (r *InMemoryMovieRepository) GetByID(id string) (*models.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.movies[id]
	if !ok {
		return nil, ErrMovieNotFound
	}
	return m, nil
}

func (r *InMemoryMovieRepository) GetAll() ([]*models.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Movie, 0, len(r.movies))
	for _, m := range r.movies {
		result = append(result, m)
	}
	return result, nil
}

func (r *InMemoryMovieRepository) SearchByTitle(title string) ([]*models.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Movie
	for _, m := range r.movies {
		if containsIgnoreCase(m.Title, title) {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryMovieRepository) SearchByGenre(genre models.Genre) ([]*models.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Movie
	for _, m := range r.movies {
		if m.Genre == genre {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *InMemoryMovieRepository) SearchByLanguage(language string) ([]*models.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Movie
	for _, m := range r.movies {
		if m.Language == language {
			result = append(result, m)
		}
	}
	return result, nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
