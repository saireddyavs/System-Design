package repositories

import (
	"errors"
	"ride-sharing-service/internal/models"
	"sync"
)

var ErrRideNotFound = errors.New("ride not found")

// InMemoryRideRepository implements RideRepository with thread-safe in-memory storage
type InMemoryRideRepository struct {
	rides map[string]*models.Ride
	mu    sync.RWMutex
}

// NewInMemoryRideRepository creates a new in-memory ride repository
func NewInMemoryRideRepository() *InMemoryRideRepository {
	return &InMemoryRideRepository{
		rides: make(map[string]*models.Ride),
	}
}

// Create adds a new ride
func (r *InMemoryRideRepository) Create(ride *models.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rides[ride.ID] = ride
	return nil
}

// GetByID retrieves a ride by ID
func (r *InMemoryRideRepository) GetByID(id string) (*models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ride, ok := r.rides[id]
	if !ok {
		return nil, ErrRideNotFound
	}
	return ride, nil
}

// Update updates an existing ride
func (r *InMemoryRideRepository) Update(ride *models.Ride) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rides[ride.ID]; !ok {
		return ErrRideNotFound
	}
	r.rides[ride.ID] = ride
	return nil
}

// GetActiveRidesByRider returns rides that are not completed/cancelled for a rider
func (r *InMemoryRideRepository) GetActiveRidesByRider(riderID string) ([]*models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Ride
	for _, ride := range r.rides {
		if ride.RiderID == riderID && ride.GetStatus() != models.RideStatusCompleted && ride.GetStatus() != models.RideStatusCancelled {
			result = append(result, ride)
		}
	}
	return result, nil
}

// GetActiveRidesByDriver returns rides that are not completed/cancelled for a driver
func (r *InMemoryRideRepository) GetActiveRidesByDriver(driverID string) ([]*models.Ride, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Ride
	for _, ride := range r.rides {
		if ride.DriverID == driverID && ride.GetStatus() != models.RideStatusCompleted && ride.GetStatus() != models.RideStatusCancelled {
			result = append(result, ride)
		}
	}
	return result, nil
}

// CountActiveRequests returns count of rides in requested or driver_assigned state
func (r *InMemoryRideRepository) CountActiveRequests() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, ride := range r.rides {
		status := ride.GetStatus()
		if status == models.RideStatusRequested || status == models.RideStatusDriverAssigned || status == models.RideStatusInProgress {
			count++
		}
	}
	return count, nil
}
