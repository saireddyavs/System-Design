package repositories

import (
	"buzzer-game/internal/models"
	"sync"
)

type InMemoryPlayerRepository struct {
	mu      sync.RWMutex
	players map[string]*models.Player
	order   []string // maintains insertion order
}

func NewInMemoryPlayerRepository() *InMemoryPlayerRepository {
	return &InMemoryPlayerRepository{
		players: make(map[string]*models.Player),
	}
}

func (r *InMemoryPlayerRepository) Add(player *models.Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.players[player.ID]; exists {
		return models.ErrDuplicatePlayer
	}
	r.players[player.ID] = player
	r.order = append(r.order, player.ID)
	return nil
}

func (r *InMemoryPlayerRepository) GetByID(id string) (*models.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	player, exists := r.players[id]
	if !exists {
		return nil, models.ErrPlayerNotFound
	}
	return player, nil
}

func (r *InMemoryPlayerRepository) GetAll() []*models.Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*models.Player, 0, len(r.order))
	for _, id := range r.order {
		if p, ok := r.players[id]; ok {
			result = append(result, p)
		}
	}
	return result
}

func (r *InMemoryPlayerRepository) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.players[id]; !exists {
		return models.ErrPlayerNotFound
	}
	delete(r.players, id)
	for i, oid := range r.order {
		if oid == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return nil
}
