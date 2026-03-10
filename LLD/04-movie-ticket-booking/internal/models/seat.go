package models

// SeatCategory represents the type of seat
type SeatCategory string

const (
	SeatCategoryRegular SeatCategory = "Regular"
	SeatCategoryPremium SeatCategory = "Premium"
	SeatCategoryVIP     SeatCategory = "VIP"
)

// Seat represents a seat in a screen
type Seat struct {
	ID       string       `json:"id"`
	ScreenID string       `json:"screen_id"`
	Row      string       `json:"row"`
	Number   int          `json:"number"`
	Category SeatCategory `json:"category"`
}

// SeatID returns unique identifier for seat in a show context
func (s *Seat) SeatID() string {
	return s.ID
}
