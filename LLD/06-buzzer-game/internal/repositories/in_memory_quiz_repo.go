package repositories

import (
	"buzzer-game/internal/models"
	"sync"
)

type InMemoryQuizRepository struct {
	mu      sync.RWMutex
	quizzes map[string]*models.Quiz
}

func NewInMemoryQuizRepository() *InMemoryQuizRepository {
	return &InMemoryQuizRepository{
		quizzes: make(map[string]*models.Quiz),
	}
}

func (r *InMemoryQuizRepository) Save(quiz *models.Quiz) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quizzes[quiz.ID] = quiz
	return nil
}

func (r *InMemoryQuizRepository) GetByID(id string) (*models.Quiz, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quiz, exists := r.quizzes[id]
	if !exists {
		return nil, models.ErrQuizNotFound
	}
	return quiz, nil
}
