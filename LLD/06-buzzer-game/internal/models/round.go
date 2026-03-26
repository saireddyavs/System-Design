package models

import "github.com/google/uuid"

type Round struct {
	ID        string
	Number    int
	Questions []*Question
}

func NewRound(number int, questions []*Question) *Round {
	return &Round{
		ID:        uuid.New().String(),
		Number:    number,
		Questions: questions,
	}
}
