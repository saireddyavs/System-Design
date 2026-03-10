package repositories

import (
	"errors"
	"food-delivery-system/internal/models"
	"sync"
)

var ErrAgentNotFound = errors.New("delivery agent not found")

// InMemoryAgentRepo implements AgentRepository with thread-safe in-memory storage
type InMemoryAgentRepo struct {
	agents map[string]*models.DeliveryAgent
	mu     sync.RWMutex
}

// NewInMemoryAgentRepo creates a new in-memory agent repository
func NewInMemoryAgentRepo() *InMemoryAgentRepo {
	return &InMemoryAgentRepo{
		agents: make(map[string]*models.DeliveryAgent),
	}
}

func (r *InMemoryAgentRepo) Create(agent *models.DeliveryAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.agents[agent.ID]; exists {
		return errors.New("agent already exists")
	}
	r.agents[agent.ID] = agent
	return nil
}

func (r *InMemoryAgentRepo) GetByID(id string) (*models.DeliveryAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, exists := r.agents[id]
	if !exists {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

func (r *InMemoryAgentRepo) GetAll() ([]*models.DeliveryAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.DeliveryAgent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, a)
	}
	return result, nil
}

func (r *InMemoryAgentRepo) GetAvailableAgents() ([]*models.DeliveryAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*models.DeliveryAgent
	for _, a := range r.agents {
		if a.Status == models.AgentStatusAvailable {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *InMemoryAgentRepo) Update(agent *models.DeliveryAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.agents[agent.ID]; !exists {
		return ErrAgentNotFound
	}
	r.agents[agent.ID] = agent
	return nil
}

func (r *InMemoryAgentRepo) UpdateLocation(agentID string, location models.Location) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	agent, exists := r.agents[agentID]
	if !exists {
		return ErrAgentNotFound
	}
	agent.UpdateLocation(location)
	return nil
}
