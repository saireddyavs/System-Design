package interfaces

import "buzzer-game/internal/models"

// BuzzerStrategy determines which player wins the buzzer when
// multiple players press simultaneously. Default: first-come-first-served.
type BuzzerStrategy interface {
	SelectWinner(events []models.BuzzerEvent) *models.BuzzerEvent
	Name() string
}
