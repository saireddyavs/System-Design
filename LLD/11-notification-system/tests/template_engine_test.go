package tests

import (
	"testing"

	"notification-system/internal/models"
	"notification-system/internal/services"
)

func TestTemplateEngine_Render(t *testing.T) {
	engine := services.NewDefaultTemplateEngine()

	template := models.NewTemplate("t1", "Welcome", models.ChannelEmail,
		"Hello {{name}}!",
		"Welcome {{name}}, your code is {{code}}.")

	subject, body, err := engine.Render(template, map[string]string{
		"name": "John",
		"code": "ABC123",
	})

	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if subject != "Hello John!" {
		t.Errorf("Subject: got %q", subject)
	}
	if body != "Welcome John, your code is ABC123." {
		t.Errorf("Body: got %q", body)
	}
}

func TestTemplateEngine_MissingVariable_KeepsPlaceholder(t *testing.T) {
	engine := services.NewDefaultTemplateEngine()

	template := models.NewTemplate("t1", "Test", models.ChannelEmail,
		"Hi {{name}}",
		"Code: {{code}}")

	_, body, err := engine.Render(template, map[string]string{"name": "Jane"})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if body != "Code: {{code}}" {
		t.Errorf("Missing variable should keep placeholder: got %q", body)
	}
}
