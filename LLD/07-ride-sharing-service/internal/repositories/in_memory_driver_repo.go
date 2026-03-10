package repositories

import (
	"errors"
	"ride-sharing-service/internal/models"
	"sync"
)

var ErrDriverNotFound = errors.New("driver not found")

// InMemoryDriverRepository implements DriverRepository with thread-safe in-memory storage
type InMemoryDriverRepository struct {
	drivers map[string]*models.Driver
	mu      sync.RWMutex
}

// NewInMemoryDriverRepository creates a new in-memory driver repository
func NewInMemoryDriverRepository() *InMemoryDriverRepository {
	return &InMemoryDriverRepository{
		drivers: make(map[string]*models.Driver),
	}
}

// Create adds a new driver
func (r *InMemoryDriverRepository) Create(driver *models.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[driver.ID] = driver
	return nil
}

// GetByID retrieves a driver by ID
func (r *InMemoryDriverRepository) GetByID(id string) (*models.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	driver, ok := r.drivers[id]
	if !ok {
		return nil, ErrDriverNotFound
	}
	return driver, nil
}

// Update updates an existing driver
func (r *InMemoryDriverRepository) Update(driver *models.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drivers[driver.ID]; !ok {
		return ErrDriverNotFound
	}
	r.drivers[driver.ID] = driver
	return nil
}

// GetAvailableDrivers returns all available drivers
func (r *InMemoryDriverRepository) GetAvailableDrivers() ([]*models.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Driver
	for _, d := range r.drivers {
		if d.IsAvailable() && d.GetRating() >= 3.0 {
			result = append(result, d)
		}
	}
	return result, nil
}

// GetAvailableDriversNear returns available drivers within radius
func (r *InMemoryDriverRepository) GetAvailableDriversNear(location models.Location, radiusKm float64) ([]*models.Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Driver
	for _, d := range r.drivers {
		if !d.IsAvailable() || d.GetRating() < 3.0 {
			continue
		}
		if models.HaversineDistance(location, d.GetLocation()) <= radiusKm {
			result = append(result, d)
		}
	}
	return result, nil
}
