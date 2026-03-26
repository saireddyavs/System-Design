package strategies

import (
	"buzzer-game/internal/models"
	"sort"
)

// ScoreBasedLeaderboardGenerator ranks players by descending score.
type ScoreBasedLeaderboardGenerator struct{}

func NewScoreBasedLeaderboardGenerator() *ScoreBasedLeaderboardGenerator {
	return &ScoreBasedLeaderboardGenerator{}
}

func (g *ScoreBasedLeaderboardGenerator) Generate(players []*models.Player, roundNumber int) *models.Leaderboard {
	sorted := make([]*models.Player, len(players))
	copy(sorted, players)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].GetScore() == sorted[j].GetScore() {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].GetScore() > sorted[j].GetScore()
	})

	entries := make([]models.LeaderboardEntry, len(sorted))
	for i, p := range sorted {
		entries[i] = models.LeaderboardEntry{
			Rank:       i + 1,
			PlayerID:   p.ID,
			PlayerName: p.Name,
			Score:      p.GetScore(),
		}
	}

	return &models.Leaderboard{
		RoundNumber: roundNumber,
		Entries:     entries,
	}
}
