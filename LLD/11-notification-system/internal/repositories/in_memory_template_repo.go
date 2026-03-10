package repositories

import (
	"context"
	"sync"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// InMemoryTemplateRepository is a thread-safe in-memory template store
type InMemoryTemplateRepository struct {
	byID   map[string]*models.Template
	byName map[string]*models.Template
	mu     sync.RWMutex
}

// NewInMemoryTemplateRepository creates a new in-memory template repo
func NewInMemoryTemplateRepository() *InMemoryTemplateRepository {
	return &InMemoryTemplateRepository{
		byID:   make(map[string]*models.Template),
		byName: make(map[string]*models.Template),
	}
}

// GetByID retrieves a template by ID
func (r *InMemoryTemplateRepository) GetByID(ctx context.Context, id string) (*models.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

// GetByName retrieves a template by name
func (r *InMemoryTemplateRepository) GetByName(ctx context.Context, name string) (*models.Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byName[name]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

// Save stores or updates a template
func (r *InMemoryTemplateRepository) Save(ctx context.Context, template *models.Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byID[template.ID] = template
	r.byName[template.Name] = template
	return nil
}

var _ interfaces.TemplateRepository = (*InMemoryTemplateRepository)(nil)
