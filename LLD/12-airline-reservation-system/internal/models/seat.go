package models

// Seat represents a seat on a flight
type Seat struct {
	ID         string
	FlightID   string
	SeatNumber string // e.g., "12A", "1F"
	Row        int
	Column     string // A, B, C, D, E, F (window seats: A, F; aisle: C, D)
	Class      SeatClass
	Status     SeatStatus
	Price      float64
}

// NewSeat creates a new Seat instance
func NewSeat(id, flightID, seatNumber string, row int, column string, class SeatClass, price float64) *Seat {
	return &Seat{
		ID:         id,
		FlightID:   flightID,
		SeatNumber: seatNumber,
		Row:        row,
		Column:     column,
		Class:      class,
		Status:     SeatStatusAvailable,
		Price:      price,
	}
}

// IsWindow returns true if the seat is a window seat (A or F typically)
func (s *Seat) IsWindow() bool {
	return s.Column == "A" || s.Column == "F"
}

// IsAisle returns true if the seat is an aisle seat (C or D typically)
func (s *Seat) IsAisle() bool {
	return s.Column == "C" || s.Column == "D"
}

// IsAvailable returns true if the seat can be booked
func (s *Seat) IsAvailable() bool {
	return s.Status == SeatStatusAvailable
}
