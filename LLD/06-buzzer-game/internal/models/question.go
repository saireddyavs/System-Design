package models

import "github.com/google/uuid"

type Option struct {
	Label string // "A", "B", "C", "D"
	Text  string
}

type Question struct {
	ID             string
	Text           string
	Options        []Option
	CorrectOption  string // label of the correct option, e.g. "B"
	TimeLimitSec   int    // how long the question stays open (default 30)
	AnswerTimeSec  int    // how long a player has after buzzing (default 5)
	PointsCorrect  int    // points for correct answer (default +3)
	PenaltyWrong   int    // penalty for wrong answer (default -1)
	PenaltyTimeout int    // penalty for answer timeout after buzzing (default 0)
}

func NewQuestion(text string, options []Option, correctOption string) *Question {
	return &Question{
		ID:             uuid.New().String(),
		Text:           text,
		Options:        options,
		CorrectOption:  correctOption,
		TimeLimitSec:   30,
		AnswerTimeSec:  5,
		PointsCorrect:  3,
		PenaltyWrong:   -1,
		PenaltyTimeout: 0,
	}
}

func (q *Question) IsCorrect(answer string) bool {
	return q.CorrectOption == answer
}
