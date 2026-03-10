package services

import (
	"fmt"
	"parking-lot-system/internal/interfaces"
	"parking-lot-system/internal/models"
	"sync"
	"strings"
)

// ParkingService orchestrates park/unpark operations.
// DIP: Depends on interfaces (ParkingStrategy), not concrete types.
// SRP: Only responsible for parking operations.
type ParkingService struct {
	mu              sync.RWMutex
	lot             *models.ParkingLot
	strategy        interfaces.ParkingStrategy
	tickets         map[string]*models.Ticket
	ticketCounter   int
}

// NewParkingService creates a parking service with the given strategy.
func NewParkingService(lot *models.ParkingLot, strategy interfaces.ParkingStrategy) *ParkingService {
	return &ParkingService{
		lot:      lot,
		strategy: strategy,
		tickets:  make(map[string]*models.Ticket),
	}
}

// Park finds a spot, parks the vehicle, and returns a ticket.
// Returns error if no spot available or vehicle already parked.
func (s *ParkingService) Park(vehicle models.Vehicle) (*models.Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if vehicle already parked
	for _, t := range s.tickets {
		if t.LicensePlate == vehicle.GetLicensePlate() {
			return nil, fmt.Errorf("vehicle %s is already parked", vehicle.GetLicensePlate())
		}
	}

	// Find best spot across all levels
	var bestSpot *models.ParkingSpot
	var bestLevel *models.ParkingLevel

	for _, level := range s.lot.GetLevels() {
		spots := level.GetAvailableSpots(vehicle)
		if len(spots) == 0 {
			continue
		}
		spot := s.strategy.FindSpot(vehicle, spots)
		if spot != nil {
			bestSpot = spot
			bestLevel = level
			break
		}
	}

	if bestSpot == nil || bestLevel == nil {
		return nil, fmt.Errorf("no available spot for vehicle type %s", vehicle.GetType().String())
	}

	if !bestSpot.Park(vehicle) {
		return nil, fmt.Errorf("failed to park vehicle at spot %s", bestSpot.ID)
	}

	s.ticketCounter++
	ticketID := fmt.Sprintf("TKT-%d", s.ticketCounter)
	ticket := models.NewTicket(ticketID, vehicle, bestSpot.ID, bestLevel.ID)
	s.tickets[ticketID] = ticket

	return ticket, nil
}

// Unpark finds the ticket by ID or license plate, unparks the vehicle, and returns it.
func (s *ParkingService) Unpark(ticketIDOrLicense string) (*models.Ticket, models.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ticket *models.Ticket
	for id, t := range s.tickets {
		if id == ticketIDOrLicense || strings.EqualFold(t.LicensePlate, ticketIDOrLicense) {
			ticket = t
			delete(s.tickets, id)
			break
		}
	}

	if ticket == nil {
		return nil, nil, fmt.Errorf("ticket or license plate not found: %s", ticketIDOrLicense)
	}

	level := s.lot.GetLevel(ticket.LevelID)
	if level == nil {
		return nil, nil, fmt.Errorf("level %s not found", ticket.LevelID)
	}

	spot := level.GetSpot(ticket.SpotID)
	if spot == nil {
		return nil, nil, fmt.Errorf("spot %s not found", ticket.SpotID)
	}

	vehicle, _ := spot.Unpark()
	if vehicle == nil {
		return nil, nil, fmt.Errorf("spot %s was already empty", ticket.SpotID)
	}

	return ticket, vehicle, nil
}

// GetTicket returns a ticket by ID or license plate.
func (s *ParkingService) GetTicket(ticketIDOrLicense string) (*models.Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, t := range s.tickets {
		if id == ticketIDOrLicense || strings.EqualFold(t.LicensePlate, ticketIDOrLicense) {
			return t, true
		}
	}
	return nil, false
}

// GetAvailableSpotsCount returns total available spots for the vehicle type across all levels.
func (s *ParkingService) GetAvailableSpotsCount(vehicle models.Vehicle) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, level := range s.lot.GetLevels() {
		count += level.CountAvailableSpots(vehicle)
	}
	return count
}
