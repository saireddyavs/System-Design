package models

import "time"

// Flight represents a scheduled flight
type Flight struct {
	ID             string
	FlightNumber   string
	Origin         string      // Airport IATA code
	Destination    string      // Airport IATA code
	DepartureTime  time.Time
	ArrivalTime    time.Time
	Aircraft       string
	Seats          []*Seat
	Status         FlightStatus
	BasePrice      float64
}

// NewFlight creates a new Flight with empty seats
func NewFlight(id, flightNumber, origin, destination, aircraft string, departureTime, arrivalTime time.Time, basePrice float64) *Flight {
	return &Flight{
		ID:            id,
		FlightNumber:  flightNumber,
		Origin:        origin,
		Destination:   destination,
		DepartureTime: departureTime,
		ArrivalTime:   arrivalTime,
		Aircraft:      aircraft,
		Seats:         make([]*Seat, 0),
		Status:        FlightStatusScheduled,
		BasePrice:     basePrice,
	}
}
