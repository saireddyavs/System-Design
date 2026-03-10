package interfaces

import "notification-system/internal/models"

// TemplateEngine renders templates with variable substitution
type TemplateEngine interface {
	// Render replaces {{variable}} placeholders with values from the map
	Render(template *models.Template, variables map[string]string) (subject, body string, err error)
}
