package models

import "time"

// Booking represents a flight booking made by a passenger
type Booking struct {
	ID          string
	PassengerID string
	FlightID    string
	SeatIDs     []string
	TotalAmount float64
	Status      BookingStatus
	BookingRef  string // Unique booking reference (e.g., "ABC123")
	CreatedAt   time.Time
}

// NewBooking creates a new Booking instance
func NewBooking(id, passengerID, flightID string, seatIDs []string, totalAmount float64, bookingRef string) *Booking {
	return &Booking{
		ID:          id,
		PassengerID: passengerID,
		FlightID:    flightID,
		SeatIDs:     seatIDs,
		TotalAmount: totalAmount,
		Status:      BookingStatusConfirmed,
		BookingRef:  bookingRef,
		CreatedAt:   time.Now(),
	}
}

// GetBaggageAllowance returns total baggage allowance based on seat classes
func (b *Booking) GetBaggageAllowance(seats []*Seat) int {
	total := 0
	seatMap := make(map[string]*Seat)
	for _, s := range seats {
		seatMap[s.ID] = s
	}
	for _, seatID := range b.SeatIDs {
		if seat, ok := seatMap[seatID]; ok {
			total += seat.Class.BaggageAllowance()
		}
	}
	return total
}
