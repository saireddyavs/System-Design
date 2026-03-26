package interfaces

import "buzzer-game/internal/models"

// LeaderboardGenerator builds a leaderboard from the current player scores.
type LeaderboardGenerator interface {
	Generate(players []*models.Player, roundNumber int) *models.Leaderboard
}
