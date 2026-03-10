package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"time"
)

// SearchCriteria defines search parameters
type SearchCriteria struct {
	Origin      string
	Destination string
	Date        time.Time
	Class       models.SeatClass // optional filter
}

// SearchResult represents a flight with availability info
type SearchResult struct {
	Flight           *models.Flight
	AvailableSeats   int
	AvailableByClass map[models.SeatClass]int
	MinPrice         float64
}

// SearchService handles flight search operations
type SearchService struct {
	flightRepo interfaces.FlightRepository
}

// NewSearchService creates a new search service
func NewSearchService(flightRepo interfaces.FlightRepository) *SearchService {
	return &SearchService{flightRepo: flightRepo}
}

// SearchFlights searches flights by route, date, and optionally class
func (s *SearchService) SearchFlights(criteria SearchCriteria) ([]*SearchResult, error) {
	flights, err := s.flightRepo.SearchByRoute(criteria.Origin, criteria.Destination, criteria.Date)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(flights))
	for _, flight := range flights {
		if flight.Status == models.FlightStatusCancelled {
			continue
		}

		availableByClass := make(map[models.SeatClass]int)
		var availableSeats []*models.Seat
		minPrice := 0.0

		for _, seat := range flight.Seats {
			if seat.IsAvailable() {
				availableByClass[seat.Class]++
				if criteria.Class == "" || seat.Class == criteria.Class {
					availableSeats = append(availableSeats, seat)
					if minPrice == 0 || seat.Price < minPrice {
						minPrice = seat.Price
					}
				}
			}
		}

		totalAvailable := len(availableSeats)
		if criteria.Class != "" {
			totalAvailable = availableByClass[criteria.Class]
		}

		if totalAvailable > 0 {
			results = append(results, &SearchResult{
				Flight:           flight,
				AvailableSeats:   totalAvailable,
				AvailableByClass: availableByClass,
				MinPrice:         minPrice,
			})
		}
	}
	return results, nil
}
