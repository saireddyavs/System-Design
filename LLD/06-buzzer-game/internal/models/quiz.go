package models

import "github.com/google/uuid"

type Quiz struct {
	ID     string
	Title  string
	Rounds []*Round
}

func NewQuiz(title string, rounds []*Round) *Quiz {
	return &Quiz{
		ID:     uuid.New().String(),
		Title:  title,
		Rounds: rounds,
	}
}
