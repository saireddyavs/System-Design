package interfaces

import (
	"context"

	"notification-system/internal/models"
)

// TemplateRepository defines operations for template storage
type TemplateRepository interface {
	GetByID(ctx context.Context, id string) (*models.Template, error)
	GetByName(ctx context.Context, name string) (*models.Template, error)
	Save(ctx context.Context, template *models.Template) error
}
