package repositories

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InMemoryCategoryRepo implements CategoryRepository with thread-safe in-memory storage
type InMemoryCategoryRepo struct {
	mu         sync.RWMutex
	categories map[string]*models.Category
}

// NewInMemoryCategoryRepo creates a new in-memory category repository
func NewInMemoryCategoryRepo() *InMemoryCategoryRepo {
	return &InMemoryCategoryRepo{
		categories: make(map[string]*models.Category),
	}
}

var _ interfaces.CategoryRepository = (*InMemoryCategoryRepo)(nil)

func (r *InMemoryCategoryRepo) Create(ctx context.Context, category *models.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.categories[category.ID]; exists {
		return ErrAlreadyExists
	}
	r.categories[category.ID] = category
	return nil
}

func (r *InMemoryCategoryRepo) GetByID(ctx context.Context, id string) (*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.categories[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (r *InMemoryCategoryRepo) List(ctx context.Context, parentID *string) ([]*models.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Category
	for _, c := range r.categories {
		if parentID == nil && c.ParentID == nil {
			result = append(result, c)
		} else if parentID != nil && c.ParentID != nil && *c.ParentID == *parentID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *InMemoryCategoryRepo) Update(ctx context.Context, category *models.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.categories[category.ID]; !ok {
		return ErrNotFound
	}
	r.categories[category.ID] = category
	return nil
}

func (r *InMemoryCategoryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.categories[id]; !ok {
		return ErrNotFound
	}
	delete(r.categories, id)
	return nil
}
