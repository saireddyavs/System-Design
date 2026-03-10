package models

import "time"

// Rating represents a rating given by one user to another after a ride
type Rating struct {
	ID        string
	RideID    string
	FromUserID string // User who gave the rating
	ToUserID   string // User who received the rating (driver or rider)
	Score     float64 // 1-5
	Comment   string
	CreatedAt time.Time
}
