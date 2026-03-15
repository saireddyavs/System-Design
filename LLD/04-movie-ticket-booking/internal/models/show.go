package models

import "time"

// SeatStatus represents availability of a seat
type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "available"
	SeatStatusBooked    SeatStatus = "booked"
)

// Show represents a scheduled movie show
type Show struct {
	ID             string               `json:"id"`
	MovieID        string               `json:"movie_id"`
	ScreenID       string               `json:"screen_id"`
	TheatreID      string               `json:"theatre_id"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	SeatStatusMap  map[string]SeatStatus `json:"-"` // seatID -> status
	BasePrice      float64              `json:"base_price"`
}
