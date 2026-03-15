package services

import (
	"context"

	"notification-system/internal/interfaces"
)

// TemplateService handles template operations
type TemplateService struct {
	repo   interfaces.TemplateRepository
	engine interfaces.TemplateEngine
}

// NewTemplateService creates a new template service
func NewTemplateService(repo interfaces.TemplateRepository, engine interfaces.TemplateEngine) *TemplateService {
	return &TemplateService{
		repo:   repo,
		engine: engine,
	}
}

// RenderTemplate renders a template with the given variables
func (s *TemplateService) RenderTemplate(ctx context.Context, templateID string, variables map[string]string) (subject, body string, err error) {
	template, err := s.repo.GetByID(ctx, templateID)
	if err != nil {
		return "", "", err
	}
	return s.engine.Render(template, variables)
}
