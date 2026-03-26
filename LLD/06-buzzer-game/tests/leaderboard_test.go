package tests

import (
	"buzzer-game/internal/models"
	"buzzer-game/internal/strategies"
	"testing"
)

func TestLeaderboardGenerator_SortsByScoreDescending(t *testing.T) {
	gen := strategies.NewScoreBasedLeaderboardGenerator()

	p1 := models.NewPlayer("Alice")
	p1.AddScore(5)
	p2 := models.NewPlayer("Bob")
	p2.AddScore(10)
	p3 := models.NewPlayer("Charlie")
	p3.AddScore(3)

	lb := gen.Generate([]*models.Player{p1, p2, p3}, 1)

	if len(lb.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(lb.Entries))
	}
	if lb.Entries[0].PlayerName != "Bob" || lb.Entries[0].Score != 10 {
		t.Errorf("expected Bob(10) at rank 1, got %s(%d)", lb.Entries[0].PlayerName, lb.Entries[0].Score)
	}
	if lb.Entries[1].PlayerName != "Alice" || lb.Entries[1].Score != 5 {
		t.Errorf("expected Alice(5) at rank 2, got %s(%d)", lb.Entries[1].PlayerName, lb.Entries[1].Score)
	}
	if lb.Entries[2].PlayerName != "Charlie" || lb.Entries[2].Score != 3 {
		t.Errorf("expected Charlie(3) at rank 3, got %s(%d)", lb.Entries[2].PlayerName, lb.Entries[2].Score)
	}
}

func TestLeaderboardGenerator_TiebreakByName(t *testing.T) {
	gen := strategies.NewScoreBasedLeaderboardGenerator()

	p1 := models.NewPlayer("Charlie")
	p1.AddScore(5)
	p2 := models.NewPlayer("Alice")
	p2.AddScore(5)

	lb := gen.Generate([]*models.Player{p1, p2}, 1)

	if lb.Entries[0].PlayerName != "Alice" {
		t.Errorf("expected Alice first (alphabetical tiebreak), got %s", lb.Entries[0].PlayerName)
	}
}

func TestLeaderboardGenerator_EmptyPlayers(t *testing.T) {
	gen := strategies.NewScoreBasedLeaderboardGenerator()
	lb := gen.Generate([]*models.Player{}, 1)

	if len(lb.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(lb.Entries))
	}
}
