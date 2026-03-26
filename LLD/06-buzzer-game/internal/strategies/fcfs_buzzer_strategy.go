package strategies

import "buzzer-game/internal/models"

// FCFSBuzzerStrategy selects the first player who pressed the buzzer.
type FCFSBuzzerStrategy struct{}

func NewFCFSBuzzerStrategy() *FCFSBuzzerStrategy {
	return &FCFSBuzzerStrategy{}
}

func (s *FCFSBuzzerStrategy) SelectWinner(events []models.BuzzerEvent) *models.BuzzerEvent {
	if len(events) == 0 {
		return nil
	}
	winner := events[0]
	for i := 1; i < len(events); i++ {
		if events[i].PressedAt.Before(winner.PressedAt) {
			winner = events[i]
		}
	}
	return &winner
}

func (s *FCFSBuzzerStrategy) Name() string {
	return "First-Come-First-Served"
}
