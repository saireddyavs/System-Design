package models

import "errors"

var (
	ErrPlayerNotFound    = errors.New("player not found")
	ErrQuizNotFound      = errors.New("quiz not found")
	ErrGameNotStarted    = errors.New("game has not started")
	ErrGameAlreadyActive = errors.New("game is already active")
	ErrBuzzerLocked      = errors.New("buzzer is locked for this player")
	ErrNotBuzzerHolder   = errors.New("player does not hold the buzzer")
	ErrQuestionTimedOut  = errors.New("question time has expired")
	ErrAnswerTimedOut    = errors.New("answer time has expired")
	ErrNoBuzzerPressed   = errors.New("no buzzer has been pressed")
	ErrInvalidOption     = errors.New("invalid answer option")
	ErrDuplicatePlayer   = errors.New("player already registered")
)
