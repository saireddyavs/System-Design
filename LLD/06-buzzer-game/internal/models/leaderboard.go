package models

// LeaderboardEntry represents a single player's standing.
type LeaderboardEntry struct {
	Rank       int
	PlayerID   string
	PlayerName string
	Score      int
}

// Leaderboard holds the ranked entries for a round or entire game.
type Leaderboard struct {
	RoundNumber int // 0 means overall/final
	Entries     []LeaderboardEntry
}
