package repositories

import (
	"food-delivery-system/internal/models"
	"sync"
)

// InMemoryRatingRepo implements RatingRepository
type InMemoryRatingRepo struct {
	ratings []*models.Rating
	mu      sync.RWMutex
}

// NewInMemoryRatingRepo creates a new in-memory rating repository
func NewInMemoryRatingRepo() *InMemoryRatingRepo {
	return &InMemoryRatingRepo{
		ratings: []*models.Rating{},
	}
}

func (r *InMemoryRatingRepo) Create(rating *models.Rating) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ratings = append(r.ratings, rating)
	return nil
}

func (r *InMemoryRatingRepo) GetByTargetID(targetID string, ratingType models.RatingType) ([]*models.Rating, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.Rating
	for _, rating := range r.ratings {
		if rating.TargetID == targetID && rating.Type == ratingType {
			result = append(result, rating)
		}
	}
	return result, nil
}

func (r *InMemoryRatingRepo) GetAverageRating(targetID string, ratingType models.RatingType) (float64, error) {
	ratings, err := r.GetByTargetID(targetID, ratingType)
	if err != nil || len(ratings) == 0 {
		return 0, nil
	}
	var sum float64
	for _, r := range ratings {
		sum += r.Score
	}
	return sum / float64(len(ratings)), nil
}
