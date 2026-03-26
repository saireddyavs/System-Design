package models

import (
	"sync"

	"github.com/google/uuid"
)

type Player struct {
	ID   string
	Name string

	mu    sync.RWMutex
	score int
}

func NewPlayer(name string) *Player {
	return &Player{
		ID:   uuid.New().String(),
		Name: name,
	}
}

func (p *Player) GetScore() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.score
}

func (p *Player) AddScore(points int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.score += points
}

func (p *Player) ResetScore() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.score = 0
}
