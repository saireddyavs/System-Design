package services

import (
	"buzzer-game/internal/interfaces"
	"buzzer-game/internal/models"
	"sync"
	"time"
)

// BuzzerService manages the buzzer lifecycle for a single question.
// It collects buzzer presses, picks a winner via strategy, and
// tracks which players are excluded (already failed this question).
type BuzzerService struct {
	strategy interfaces.BuzzerStrategy

	mu             sync.Mutex
	currentQID     string
	buzzerEvents   []models.BuzzerEvent
	excludedPlayers map[string]bool
	buzzerHolder   string // playerID who currently holds the buzzer
	buzzerDeadline time.Time
	isLocked       bool
}

func NewBuzzerService(strategy interfaces.BuzzerStrategy) *BuzzerService {
	return &BuzzerService{
		strategy:        strategy,
		excludedPlayers: make(map[string]bool),
	}
}

// OpenForQuestion resets the buzzer state for a new question.
func (bs *BuzzerService) OpenForQuestion(questionID string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.currentQID = questionID
	bs.buzzerEvents = nil
	bs.excludedPlayers = make(map[string]bool)
	bs.buzzerHolder = ""
	bs.isLocked = false
}

// ExcludePlayer marks a player as unable to buzz for the current question.
func (bs *BuzzerService) ExcludePlayer(playerID string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.excludedPlayers[playerID] = true
}

// IsExcluded checks if a player is locked out from buzzing.
func (bs *BuzzerService) IsExcluded(playerID string) bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.excludedPlayers[playerID]
}

// PressBuzzer records a buzzer press. Returns error if the player is excluded.
func (bs *BuzzerService) PressBuzzer(playerID string) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.excludedPlayers[playerID] {
		return models.ErrBuzzerLocked
	}

	bs.buzzerEvents = append(bs.buzzerEvents, models.BuzzerEvent{
		PlayerID:   playerID,
		QuestionID: bs.currentQID,
		PressedAt:  time.Now(),
	})
	return nil
}

// ResolveBuzzer picks the winning buzzer press and grants the holder
// answerTimeSec seconds to answer. Returns the winning player ID.
func (bs *BuzzerService) ResolveBuzzer(answerTimeSec int) (string, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if len(bs.buzzerEvents) == 0 {
		return "", models.ErrNoBuzzerPressed
	}

	winner := bs.strategy.SelectWinner(bs.buzzerEvents)
	if winner == nil {
		return "", models.ErrNoBuzzerPressed
	}

	bs.buzzerHolder = winner.PlayerID
	bs.buzzerDeadline = time.Now().Add(time.Duration(answerTimeSec) * time.Second)
	bs.buzzerEvents = nil
	bs.isLocked = true

	return winner.PlayerID, nil
}

// GetBuzzerHolder returns the current holder and whether the answer window is still open.
func (bs *BuzzerService) GetBuzzerHolder() (string, bool) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	if bs.buzzerHolder == "" {
		return "", false
	}
	return bs.buzzerHolder, time.Now().Before(bs.buzzerDeadline)
}

// ClearBuzzerHolder clears the current holder after answer resolution.
func (bs *BuzzerService) ClearBuzzerHolder() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.buzzerHolder = ""
	bs.isLocked = false
}

// ReopenBuzzer allows remaining (non-excluded) players to buzz again.
func (bs *BuzzerService) ReopenBuzzer() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.buzzerEvents = nil
	bs.buzzerHolder = ""
	bs.isLocked = false
}

// GetExcludedCount returns how many players are excluded.
func (bs *BuzzerService) GetExcludedCount() int {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return len(bs.excludedPlayers)
}
