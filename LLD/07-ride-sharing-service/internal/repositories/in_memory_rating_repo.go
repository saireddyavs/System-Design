package repositories

import (
	"ride-sharing-service/internal/models"
	"sync"
)

// InMemoryRatingRepository implements RatingRepository with thread-safe in-memory storage
type InMemoryRatingRepository struct {
	ratings map[string][]*models.Rating // userID -> ratings received
	mu      sync.RWMutex
}

// NewInMemoryRatingRepository creates a new in-memory rating repository
func NewInMemoryRatingRepository() *InMemoryRatingRepository {
	return &InMemoryRatingRepository{
		ratings: make(map[string][]*models.Rating),
	}
}

// Create adds a new rating
func (r *InMemoryRatingRepository) Create(rating *models.Rating) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ratings[rating.ToUserID] = append(r.ratings[rating.ToUserID], rating)
	return nil
}

// GetRatingsForUser returns all ratings received by a user
func (r *InMemoryRatingRepository) GetRatingsForUser(userID string) ([]*models.Rating, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ratings[userID], nil
}

// GetAverageRating returns average rating and count for a user
func (r *InMemoryRatingRepository) GetAverageRating(userID string) (float64, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userRatings := r.ratings[userID]
	if len(userRatings) == 0 {
		return 0, 0, nil
	}
	var sum float64
	for _, rating := range userRatings {
		sum += rating.Score
	}
	return sum / float64(len(userRatings)), len(userRatings), nil
}
