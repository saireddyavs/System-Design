package repositories

import (
	"errors"
	"movie-ticket-booking/internal/models"
	"sync"
)

var ErrShowNotFound = errors.New("show not found")

// InMemoryShowRepository implements ShowRepository with per-show locking for concurrency
type InMemoryShowRepository struct {
	shows      map[string]*models.Show
	showLocks  map[string]*sync.Mutex
	globalLock sync.RWMutex
	mu         sync.RWMutex
}

// NewInMemoryShowRepository creates a new in-memory show repository
func NewInMemoryShowRepository() *InMemoryShowRepository {
	return &InMemoryShowRepository{
		shows:     make(map[string]*models.Show),
		showLocks: make(map[string]*sync.Mutex),
	}
}

func (r *InMemoryShowRepository) getShowLock(showID string) *sync.Mutex {
	r.globalLock.Lock()
	defer r.globalLock.Unlock()
	if lock, ok := r.showLocks[showID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	r.showLocks[showID] = lock
	return lock
}

// LockShow acquires exclusive lock on a show for seat booking (pessimistic locking)
func (r *InMemoryShowRepository) LockShow(showID string) *sync.Mutex {
	return r.getShowLock(showID)
}

func (r *InMemoryShowRepository) Create(show *models.Show) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if show.SeatStatusMap == nil {
		show.SeatStatusMap = make(map[string]models.SeatStatus)
	}
	r.shows[show.ID] = show
	r.showLocks[show.ID] = &sync.Mutex{}
	return nil
}

func (r *InMemoryShowRepository) GetByID(id string) (*models.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.shows[id]
	if !ok {
		return nil, ErrShowNotFound
	}
	return s, nil
}

func (r *InMemoryShowRepository) GetByMovieID(movieID string) ([]*models.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Show
	for _, s := range r.shows {
		if s.MovieID == movieID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *InMemoryShowRepository) GetByTheatreID(theatreID string) ([]*models.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Show
	for _, s := range r.shows {
		if s.TheatreID == theatreID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *InMemoryShowRepository) GetByScreenID(screenID string) ([]*models.Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Show
	for _, s := range r.shows {
		if s.ScreenID == screenID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *InMemoryShowRepository) Update(show *models.Show) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.shows[show.ID]; !ok {
		return ErrShowNotFound
	}
	r.shows[show.ID] = show
	return nil
}

// UpdateSeats atomically updates show seats with per-show pessimistic locking
func (r *InMemoryShowRepository) UpdateSeats(showID string, updateFn func(*models.Show) error) error {
	lock := r.getShowLock(showID)
	lock.Lock()
	defer lock.Unlock()

	r.mu.RLock()
	show, ok := r.shows[showID]
	r.mu.RUnlock()
	if !ok {
		return ErrShowNotFound
	}

	if err := updateFn(show); err != nil {
		return err
	}

	r.mu.Lock()
	r.shows[showID] = show
	r.mu.Unlock()
	return nil
}
