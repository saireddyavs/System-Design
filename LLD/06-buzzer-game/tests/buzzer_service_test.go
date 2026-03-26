package tests

import (
	"buzzer-game/internal/models"
	"buzzer-game/internal/services"
	"buzzer-game/internal/strategies"
	"testing"
	"time"
)

func TestPressBuzzer_ExcludedPlayer(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")
	bs.ExcludePlayer("player-1")

	err := bs.PressBuzzer("player-1")
	if err != models.ErrBuzzerLocked {
		t.Errorf("expected ErrBuzzerLocked, got %v", err)
	}
}

func TestPressBuzzer_AllowedPlayer(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")

	err := bs.PressBuzzer("player-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestResolveBuzzer_FirstPressWins(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")

	_ = bs.PressBuzzer("player-1")
	time.Sleep(1 * time.Millisecond)
	_ = bs.PressBuzzer("player-2")

	winnerID, err := bs.ResolveBuzzer(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winnerID != "player-1" {
		t.Errorf("expected player-1 to win, got %s", winnerID)
	}
}

func TestResolveBuzzer_NoPresses(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")

	_, err := bs.ResolveBuzzer(5)
	if err != models.ErrNoBuzzerPressed {
		t.Errorf("expected ErrNoBuzzerPressed, got %v", err)
	}
}

func TestBuzzerHolder_WithinTimeWindow(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")

	_ = bs.PressBuzzer("player-1")
	_, _ = bs.ResolveBuzzer(5)

	holder, withinTime := bs.GetBuzzerHolder()
	if holder != "player-1" {
		t.Errorf("expected player-1 as holder, got %s", holder)
	}
	if !withinTime {
		t.Error("expected to be within time window")
	}
}

func TestExcludeAndReopen(t *testing.T) {
	bs := services.NewBuzzerService(strategies.NewFCFSBuzzerStrategy())
	bs.OpenForQuestion("q1")

	bs.ExcludePlayer("player-1")
	bs.ReopenBuzzer()

	if !bs.IsExcluded("player-1") {
		t.Error("player-1 should still be excluded after reopen")
	}

	err := bs.PressBuzzer("player-2")
	if err != nil {
		t.Errorf("player-2 should be able to buzz, got %v", err)
	}
}
