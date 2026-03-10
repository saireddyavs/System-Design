package models

// Screen represents a screen/hall in a theatre
type Screen struct {
	ID            string `json:"id"`
	TheatreID     string `json:"theatre_id"`
	Name          string `json:"name"`
	TotalCapacity int    `json:"total_capacity"`
	Seats         []Seat `json:"seats"`
}
