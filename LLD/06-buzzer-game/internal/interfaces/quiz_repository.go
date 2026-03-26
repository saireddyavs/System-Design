package interfaces

import "buzzer-game/internal/models"

// QuizRepository abstracts quiz storage.
type QuizRepository interface {
	Save(quiz *models.Quiz) error
	GetByID(id string) (*models.Quiz, error)
}
