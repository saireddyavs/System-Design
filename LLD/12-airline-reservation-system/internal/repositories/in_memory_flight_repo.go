package repositories

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
	"sync"
	"time"
)

var (
	ErrFlightNotFound = errors.New("flight not found")
)

// InMemoryFlightRepository implements FlightRepository with in-memory storage (thread-safe)
type InMemoryFlightRepository struct {
	flights map[string]*models.Flight
	mu     sync.RWMutex
}

// NewInMemoryFlightRepository creates a new in-memory flight repository
func NewInMemoryFlightRepository() interfaces.FlightRepository {
	return &InMemoryFlightRepository{
		flights: make(map[string]*models.Flight),
	}
}

func (r *InMemoryFlightRepository) Create(flight *models.Flight) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flights[flight.ID]; exists {
		return errors.New("flight already exists")
	}

	// Deep copy to avoid external mutation
	flightCopy := copyFlight(flight)
	r.flights[flight.ID] = flightCopy
	return nil
}

func (r *InMemoryFlightRepository) GetByID(id string) (*models.Flight, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	flight, exists := r.flights[id]
	if !exists {
		return nil, ErrFlightNotFound
	}
	return copyFlight(flight), nil
}

func (r *InMemoryFlightRepository) Update(flight *models.Flight) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flights[flight.ID]; !exists {
		return ErrFlightNotFound
	}
	r.flights[flight.ID] = copyFlight(flight)
	return nil
}

func (r *InMemoryFlightRepository) SearchByRoute(origin, destination string, date time.Time) ([]*models.Flight, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dateStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dateEnd := dateStart.Add(24 * time.Hour)

	var result []*models.Flight
	for _, f := range r.flights {
		if f.Origin == origin && f.Destination == destination &&
			!f.DepartureTime.Before(dateStart) && f.DepartureTime.Before(dateEnd) &&
			f.Status != models.FlightStatusCancelled {
			result = append(result, copyFlight(f))
		}
	}
	return result, nil
}

func copyFlight(f *models.Flight) *models.Flight {
	seats := make([]*models.Seat, len(f.Seats))
	for i, s := range f.Seats {
		seatCopy := *s
		seats[i] = &seatCopy
	}
	return &models.Flight{
		ID:            f.ID,
		FlightNumber:  f.FlightNumber,
		Origin:        f.Origin,
		Destination:   f.Destination,
		DepartureTime: f.DepartureTime,
		ArrivalTime:   f.ArrivalTime,
		Aircraft:      f.Aircraft,
		Seats:         seats,
		Status:        f.Status,
		BasePrice:     f.BasePrice,
	}
}
