package interfaces

import "buzzer-game/internal/models"

// PlayerRepository abstracts player storage.
type PlayerRepository interface {
	Add(player *models.Player) error
	GetByID(id string) (*models.Player, error)
	GetAll() []*models.Player
	Remove(id string) error
}
