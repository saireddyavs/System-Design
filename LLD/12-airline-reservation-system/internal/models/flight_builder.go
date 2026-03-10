package models

import (
	"fmt"
	"time"
)

// FlightBuilder implements Builder pattern for flight creation with seats
type FlightBuilder struct {
	id            string
	flightNumber  string
	origin        string
	destination   string
	departureTime time.Time
	arrivalTime   time.Time
	aircraft      string
	basePrice     float64
	seatConfig    []seatConfig
}

type seatConfig struct {
	rows    int
	columns []string
	class   SeatClass
}

// NewFlightBuilder creates a new flight builder
func NewFlightBuilder() *FlightBuilder {
	return &FlightBuilder{
		basePrice:  100.0,
		seatConfig: make([]seatConfig, 0),
	}
}

func (b *FlightBuilder) ID(id string) *FlightBuilder {
	b.id = id
	return b
}

func (b *FlightBuilder) FlightNumber(num string) *FlightBuilder {
	b.flightNumber = num
	return b
}

func (b *FlightBuilder) Route(origin, destination string) *FlightBuilder {
	b.origin = origin
	b.destination = destination
	return b
}

func (b *FlightBuilder) Schedule(departure, arrival time.Time) *FlightBuilder {
	b.departureTime = departure
	b.arrivalTime = arrival
	return b
}

func (b *FlightBuilder) Aircraft(aircraft string) *FlightBuilder {
	b.aircraft = aircraft
	return b
}

func (b *FlightBuilder) BasePrice(price float64) *FlightBuilder {
	b.basePrice = price
	return b
}

// AddSeatSection adds a section of seats (e.g., 10 rows of ABC for Economy)
func (b *FlightBuilder) AddSeatSection(rows int, columns []string, class SeatClass) *FlightBuilder {
	b.seatConfig = append(b.seatConfig, seatConfig{rows, columns, class})
	return b
}

// Build creates the Flight with all seats configured
func (b *FlightBuilder) Build() (*Flight, error) {
	if b.id == "" || b.flightNumber == "" || b.origin == "" || b.destination == "" {
		return nil, fmt.Errorf("missing required flight fields")
	}

	flight := NewFlight(b.id, b.flightNumber, b.origin, b.destination, b.aircraft,
		b.departureTime, b.arrivalTime, b.basePrice)

	seatID := 0
	for _, config := range b.seatConfig {
		for row := 1; row <= config.rows; row++ {
			for _, col := range config.columns {
				seatNumber := fmt.Sprintf("%d%s", row, col)
				price := b.basePrice * config.class.ClassMultiplier()
				seat := NewSeat(
					fmt.Sprintf("%s-seat-%d", b.id, seatID),
					b.id,
					seatNumber,
					row,
					col,
					config.class,
					price,
				)
				flight.Seats = append(flight.Seats, seat)
				seatID++
			}
		}
	}

	return flight, nil
}
