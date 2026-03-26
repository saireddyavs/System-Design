package models

import "time"

// BuzzerEvent records a player pressing the buzzer for a question.
type BuzzerEvent struct {
	PlayerID   string
	QuestionID string
	PressedAt  time.Time
}

// AnswerEvent records a player's answer attempt after buzzing.
type AnswerEvent struct {
	PlayerID     string
	QuestionID   string
	ChosenOption string
	Result       AnswerResult
	PointsDelta  int
	AnsweredAt   time.Time
}
