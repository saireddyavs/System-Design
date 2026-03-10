package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"notification-system/internal/interfaces"
	"notification-system/internal/middleware"
	"notification-system/internal/models"
	"notification-system/internal/repositories"
	"notification-system/internal/senders"
	"notification-system/internal/services"
)

func TestSendNotification_WithTemplate(t *testing.T) {
	ctx := context.Background()

	notifRepo := repositories.NewInMemoryNotificationRepository()
	userRepo := repositories.NewInMemoryUserRepository()
	templateRepo := repositories.NewInMemoryTemplateRepository()

	user := models.NewUser("u1", "Test", "test@example.com", "+111", "token")
	_ = userRepo.Save(ctx, user)

	template := models.NewTemplate("t1", "Test", models.ChannelEmail, "Hi {{name}}", "Hello {{name}}!")
	_ = templateRepo.Save(ctx, template)

	engine := services.NewDefaultTemplateEngine()
	templateSvc := services.NewTemplateService(templateRepo, engine)
	preferenceSvc := services.NewPreferenceService(userRepo)
	rateLimiter := middleware.NewSlidingWindowRateLimiter(100, time.Hour) // High limit for tests

	senderMap := map[models.Channel]interfaces.NotificationSender{
		models.ChannelEmail: senders.NewEmailSender(),
	}

	svc := services.NewNotificationService(
		notifRepo, userRepo, senderMap, templateSvc, preferenceSvc, rateLimiter,
	)

	req := services.NewSendRequestBuilder("u1").
		WithChannel(models.ChannelEmail).
		WithTemplate("t1", map[string]string{"name": "Test"}).
		Build()

	err := svc.SendNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendNotification failed: %v", err)
	}
}

func TestSendNotification_RespectsUserPreferences(t *testing.T) {
	ctx := context.Background()

	notifRepo := repositories.NewInMemoryNotificationRepository()
	userRepo := repositories.NewInMemoryUserRepository()
	templateRepo := repositories.NewInMemoryTemplateRepository()

	user := models.NewUser("u1", "Test", "test@example.com", "+111", "token")
	user.SetChannelPreference(models.ChannelEmail, false) // Disabled
	_ = userRepo.Save(ctx, user)

	engine := services.NewDefaultTemplateEngine()
	templateSvc := services.NewTemplateService(templateRepo, engine)
	preferenceSvc := services.NewPreferenceService(userRepo)
	rateLimiter := middleware.NewSlidingWindowRateLimiter(100, time.Hour)

	senderMap := map[models.Channel]interfaces.NotificationSender{
		models.ChannelEmail: senders.NewEmailSender(),
	}

	svc := services.NewNotificationService(
		notifRepo, userRepo, senderMap, templateSvc, preferenceSvc, rateLimiter,
	)

	req := services.NewSendRequestBuilder("u1").
		WithChannel(models.ChannelEmail).
		WithContent("Subj", "Body").
		Build()

	err := svc.SendNotification(ctx, req)
	if err == nil {
		t.Fatal("Expected error when channel disabled, got nil")
	}
}

func TestSendBatch_Concurrent(t *testing.T) {
	ctx := context.Background()

	notifRepo := repositories.NewInMemoryNotificationRepository()
	userRepo := repositories.NewInMemoryUserRepository()
	templateRepo := repositories.NewInMemoryTemplateRepository()

	user := models.NewUser("u1", "Test", "test@example.com", "+111", "token")
	_ = userRepo.Save(ctx, user)

	engine := services.NewDefaultTemplateEngine()
	templateSvc := services.NewTemplateService(templateRepo, engine)
	preferenceSvc := services.NewPreferenceService(userRepo)
	rateLimiter := middleware.NewSlidingWindowRateLimiter(100, time.Hour)

	senderMap := map[models.Channel]interfaces.NotificationSender{
		models.ChannelEmail: senders.NewEmailSender(),
		models.ChannelSMS:   senders.NewSMSSender(),
	}

	svc := services.NewNotificationService(
		notifRepo, userRepo, senderMap, templateSvc, preferenceSvc, rateLimiter,
	)

	reqs := make([]*services.SendRequest, 10)
	for i := range reqs {
		ch := models.ChannelEmail
		if i%2 == 0 {
			ch = models.ChannelSMS
		}
		reqs[i] = services.NewSendRequestBuilder("u1").WithChannel(ch).WithContent("", "Batch").Build()
	}

	errors := svc.SendBatch(ctx, reqs)
	successCount := 0
	for _, e := range errors {
		if e == nil {
			successCount++
		}
	}
	if successCount != 10 {
		t.Errorf("Expected 10 successes, got %d", successCount)
	}
}

func TestObserver_ReceivesEvents(t *testing.T) {
	ctx := context.Background()

	notifRepo := repositories.NewInMemoryNotificationRepository()
	userRepo := repositories.NewInMemoryUserRepository()
	templateRepo := repositories.NewInMemoryTemplateRepository()

	user := models.NewUser("u1", "Test", "test@example.com", "+111", "token")
	_ = userRepo.Save(ctx, user)

	engine := services.NewDefaultTemplateEngine()
	templateSvc := services.NewTemplateService(templateRepo, engine)
	preferenceSvc := services.NewPreferenceService(userRepo)
	rateLimiter := middleware.NewSlidingWindowRateLimiter(100, time.Hour)

	senderMap := map[models.Channel]interfaces.NotificationSender{
		models.ChannelEmail: senders.NewEmailSender(),
	}

	svc := services.NewNotificationService(
		notifRepo, userRepo, senderMap, templateSvc, preferenceSvc, rateLimiter,
	)

	var events []services.NotificationEvent
	var mu sync.Mutex
	obs := &capturingObserver{events: &events, mu: &mu}
	svc.Subscribe(obs)

	req := services.NewSendRequestBuilder("u1").WithChannel(models.ChannelEmail).WithContent("", "Body").Build()
	_ = svc.SendNotification(ctx, req)

	mu.Lock()
	count := len(events)
	mu.Unlock()

	if count < 2 {
		t.Errorf("Expected at least 2 events (pending, sent), got %d", count)
	}
}

type capturingObserver struct {
	events *[]services.NotificationEvent
	mu     *sync.Mutex
}

func (c *capturingObserver) OnNotificationEvent(e services.NotificationEvent) {
	c.mu.Lock()
	*c.events = append(*c.events, e)
	c.mu.Unlock()
}
