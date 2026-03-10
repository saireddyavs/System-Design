package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
	"time"
)

var (
	ErrFlightNotFound   = errors.New("flight not found")
	ErrInvalidFlight    = errors.New("invalid flight data")
	ErrFlightHasBookings = errors.New("cannot cancel flight with active bookings")
)

// FlightService handles flight management operations
type FlightService struct {
	flightRepo  interfaces.FlightRepository
	bookingRepo interfaces.BookingRepository
}

// NewFlightService creates a new flight service
func NewFlightService(flightRepo interfaces.FlightRepository, bookingRepo interfaces.BookingRepository) *FlightService {
	return &FlightService{
		flightRepo:  flightRepo,
		bookingRepo: bookingRepo,
	}
}

// AddFlight adds a new flight
func (s *FlightService) AddFlight(flight *models.Flight) error {
	if flight == nil || flight.ID == "" || flight.FlightNumber == "" {
		return ErrInvalidFlight
	}
	return s.flightRepo.Create(flight)
}

// UpdateFlight updates an existing flight
func (s *FlightService) UpdateFlight(flight *models.Flight) error {
	existing, err := s.flightRepo.GetByID(flight.ID)
	if err != nil {
		return err
	}
	// Preserve seats when updating other fields
	flight.Seats = existing.Seats
	return s.flightRepo.Update(flight)
}

// CancelFlight cancels a flight (only if no active bookings)
func (s *FlightService) CancelFlight(flightID string) error {
	flight, err := s.flightRepo.GetByID(flightID)
	if err != nil {
		return err
	}

	bookings, err := s.bookingRepo.GetByFlightID(flightID)
	if err != nil {
		return err
	}

	// Check for confirmed bookings
	for _, b := range bookings {
		if b.Status == models.BookingStatusConfirmed {
			return ErrFlightHasBookings
		}
	}

	flight.Status = models.FlightStatusCancelled
	return s.flightRepo.Update(flight)
}

// GetFlight retrieves a flight by ID
func (s *FlightService) GetFlight(flightID string) (*models.Flight, error) {
	return s.flightRepo.GetByID(flightID)
}

// SearchFlights searches flights by route and date
func (s *FlightService) SearchFlights(origin, destination string, date time.Time) ([]*models.Flight, error) {
	return s.flightRepo.SearchByRoute(origin, destination, date)
}
