package models

import "time"

// RatingType represents what is being rated
type RatingType string

const (
	RatingTypeRestaurant RatingType = "restaurant"
	RatingTypeAgent      RatingType = "agent"
)

// Rating represents a rating given by a customer
type Rating struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Type       RatingType `json:"type"`
	TargetID   string    `json:"target_id"` // RestaurantID or AgentID
	Score      float64   `json:"score"`     // 1-5
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewRating creates a new rating
func NewRating(id, orderID, customerID string, ratingType RatingType, targetID string, score float64, comment string) *Rating {
	if score < 1 {
		score = 1
	}
	if score > 5 {
		score = 5
	}
	return &Rating{
		ID:         id,
		OrderID:    orderID,
		CustomerID: customerID,
		Type:       ratingType,
		TargetID:   targetID,
		Score:      score,
		Comment:    comment,
		CreatedAt:  time.Now(),
	}
}
