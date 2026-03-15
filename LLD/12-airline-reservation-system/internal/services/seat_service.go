package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"errors"
)

var (
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatNotAvailable  = errors.New("seat is not available")
	ErrInsufficientSeats = errors.New("insufficient available seats")
)

// SeatService handles seat management and assignment
type SeatService struct {
	flightRepo           interfaces.FlightRepository
	seatAssignmentStrategy interfaces.SeatAssignmentStrategy
}

// NewSeatService creates a new seat service
func NewSeatService(flightRepo interfaces.FlightRepository, strategy interfaces.SeatAssignmentStrategy) *SeatService {
	return &SeatService{
		flightRepo:            flightRepo,
		seatAssignmentStrategy: strategy,
	}
}

// GetAvailableSeats returns available seats for a flight, optionally filtered by class
func (s *SeatService) GetAvailableSeats(flightID string, class models.SeatClass) ([]*models.Seat, error) {
	flight, err := s.flightRepo.GetByID(flightID)
	if err != nil {
		return nil, err
	}

	var available []*models.Seat
	for _, seat := range flight.Seats {
		if seat.IsAvailable() {
			if class == "" || seat.Class == class {
				available = append(available, seat)
			}
		}
	}
	return available, nil
}

// AutoAssignSeats assigns seats using the configured strategy
func (s *SeatService) AutoAssignSeats(flightID string, count int, preferredClass models.SeatClass) ([]string, error) {
	available, err := s.GetAvailableSeats(flightID, preferredClass)
	if err != nil {
		return nil, err
	}
	return s.seatAssignmentStrategy.AssignSeats(available, count, preferredClass)
}

// ManualAssignSeats assigns specific seats (validates availability)
func (s *SeatService) ManualAssignSeats(flightID string, seatIDs []string) error {
	flight, err := s.flightRepo.GetByID(flightID)
	if err != nil {
		return err
	}

	seatMap := make(map[string]*models.Seat)
	for _, s := range flight.Seats {
		seatMap[s.ID] = s
	}

	for _, seatID := range seatIDs {
		seat, ok := seatMap[seatID]
		if !ok {
			return ErrSeatNotFound
		}
		if !seat.IsAvailable() {
			return ErrSeatNotAvailable
		}
	}
	return nil
}

// MarkSeatsBooked marks seats as booked (called by BookingService)
func (s *SeatService) MarkSeatsBooked(flightID string, seatIDs []string) error {
	flight, err := s.flightRepo.GetByID(flightID)
	if err != nil {
		return err
	}

	seatIDSet := make(map[string]bool)
	for _, id := range seatIDs {
		seatIDSet[id] = true
	}

	for _, seat := range flight.Seats {
		if seatIDSet[seat.ID] {
			seat.Status = models.SeatStatusBooked
		}
	}
	return s.flightRepo.Update(flight)
}

// ReleaseSeats marks seats as available (called on cancellation)
func (s *SeatService) ReleaseSeats(flightID string, seatIDs []string) error {
	flight, err := s.flightRepo.GetByID(flightID)
	if err != nil {
		return err
	}

	seatIDSet := make(map[string]bool)
	for _, id := range seatIDs {
		seatIDSet[id] = true
	}

	for _, seat := range flight.Seats {
		if seatIDSet[seat.ID] {
			seat.Status = models.SeatStatusAvailable
		}
	}
	return s.flightRepo.Update(flight)
}
