package services

import (
	"regexp"
	"strings"

	"notification-system/internal/interfaces"
	"notification-system/internal/models"
)

// DefaultTemplateEngine implements variable substitution for templates
// Uses {{variable}} placeholder syntax
type DefaultTemplateEngine struct {
	placeholderRegex *regexp.Regexp
}

// NewDefaultTemplateEngine creates a new template engine
func NewDefaultTemplateEngine() *DefaultTemplateEngine {
	return &DefaultTemplateEngine{
		placeholderRegex: regexp.MustCompile(`\{\{(\w+)\}\}`),
	}
}

// Render replaces {{variable}} placeholders with values
func (e *DefaultTemplateEngine) Render(template *models.Template, variables map[string]string) (subject, body string, err error) {
	subject = e.replacePlaceholders(template.Subject, variables)
	body = e.replacePlaceholders(template.Body, variables)
	return subject, body, nil
}

func (e *DefaultTemplateEngine) replacePlaceholders(text string, variables map[string]string) string {
	return e.placeholderRegex.ReplaceAllStringFunc(text, func(match string) string {
		key := strings.Trim(match, "{}")
		if val, ok := variables[key]; ok {
			return val
		}
		return match
	})
}

// Ensure DefaultTemplateEngine implements TemplateEngine
var _ interfaces.TemplateEngine = (*DefaultTemplateEngine)(nil)
