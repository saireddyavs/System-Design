package main

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"notification-system/internal/interfaces"
	"notification-system/internal/middleware"
	"notification-system/internal/models"
	"notification-system/internal/repositories"
	"notification-system/internal/senders"
	"notification-system/internal/services"
)

func main() {
	ctx := context.Background()

	// Initialize repositories
	notifRepo := repositories.NewInMemoryNotificationRepository()
	userRepo := repositories.NewInMemoryUserRepository()
	templateRepo := repositories.NewInMemoryTemplateRepository()

	// Seed data
	user := models.NewUser("user-1", "John Doe", "john@example.com", "+1234567890", "device-token-abc")
	_ = userRepo.Save(ctx, user)

	template := models.NewTemplate(
		"welcome-email",
		"Welcome Email",
		models.ChannelEmail,
		"Welcome {{name}}!",
		"Hi {{name}}, welcome to our platform. Your account is ready.",
	)
	_ = templateRepo.Save(ctx, template)

	// Initialize components
	templateEngine := services.NewDefaultTemplateEngine()
	templateSvc := services.NewTemplateService(templateRepo, templateEngine)
	preferenceSvc := services.NewPreferenceService(userRepo)
	rateLimiter := middleware.DefaultRateLimiter()

	// Create senders with retry decorator (Strategy + Decorator)
	emailSender := middleware.NewRetrySender(senders.NewEmailSender(), 3, 100*time.Millisecond)
	smsSender := middleware.NewRetrySender(senders.NewSMSSender(), 3, 100*time.Millisecond)
	pushSender := middleware.NewRetrySender(senders.NewPushSender(), 3, 100*time.Millisecond)

	senderMap := map[models.Channel]interfaces.NotificationSender{
		models.ChannelEmail: emailSender,
		models.ChannelSMS:   smsSender,
		models.ChannelPush:  pushSender,
	}

	notifSvc := services.NewNotificationService(
		notifRepo,
		userRepo,
		senderMap,
		templateSvc,
		preferenceSvc,
		rateLimiter,
	)

	// Add observer for logging
	notifSvc.Subscribe(&logObserver{})

	// Create worker pool with priority queue
	wp := NewNotificationWorkerPool(notifSvc, 3)
	wp.Start(ctx)
	defer wp.Stop()

	// Send single notification
	req := services.NewSendRequestBuilder("user-1").
		WithChannel(models.ChannelEmail).
		WithType(models.TypeAlert).
		WithPriority(models.PriorityHigh).
		WithTemplate("welcome-email", map[string]string{"name": "John"}).
		Build()

	if err := notifSvc.SendNotification(ctx, req); err != nil {
		log.Printf("Send error: %v", err)
	} else {
		fmt.Println("Notification sent successfully!")
	}

	// Submit to worker pool (for async processing)
	wp.Submit(req)

	// Batch send
	batchReqs := []*services.SendRequest{
		services.NewSendRequestBuilder("user-1").WithChannel(models.ChannelEmail).WithContent("Test", "Batch 1").Build(),
		services.NewSendRequestBuilder("user-1").WithChannel(models.ChannelSMS).WithContent("", "Batch 2").Build(),
	}
	errors := notifSvc.SendBatch(ctx, batchReqs)
	for i, err := range errors {
		if err != nil {
			log.Printf("Batch %d error: %v", i, err)
		}
	}

	time.Sleep(500 * time.Millisecond)
}

type logObserver struct{}

func (l *logObserver) OnNotificationEvent(e services.NotificationEvent) {
	log.Printf("[Observer] Notification %s: %s", e.NotificationID, e.Status)
}

// NotificationWorkerPool processes notifications with priority queue
type NotificationWorkerPool struct {
	service   *services.NotificationService
	workers   int
	jobQueue  *PriorityJobQueue
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// Job wraps a send request with priority
type Job struct {
	Request  *services.SendRequest
	Priority models.Priority
}

// PriorityJobQueue implements heap.Interface for priority processing
type PriorityJobQueue struct {
	jobs []*Job
	mu   sync.Mutex
}

func (q *PriorityJobQueue) Len() int { return len(q.jobs) }
func (q *PriorityJobQueue) Less(i, j int) bool {
	return q.jobs[i].Priority > q.jobs[j].Priority // Higher priority first
}
func (q *PriorityJobQueue) Swap(i, j int) {
	q.jobs[i], q.jobs[j] = q.jobs[j], q.jobs[i]
}
func (q *PriorityJobQueue) Push(x interface{}) {
	q.jobs = append(q.jobs, x.(*Job))
}
func (q *PriorityJobQueue) Pop() interface{} {
	n := len(q.jobs)
	item := q.jobs[n-1]
	q.jobs = q.jobs[:n-1]
	return item
}

// NewNotificationWorkerPool creates a worker pool
func NewNotificationWorkerPool(service *services.NotificationService, workers int) *NotificationWorkerPool {
	return &NotificationWorkerPool{
		service:  service,
		workers:  workers,
		jobQueue: &PriorityJobQueue{jobs: make([]*Job, 0)},
	}
}

// Start begins processing
func (p *NotificationWorkerPool) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	heap.Init(p.jobQueue)

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Submit adds a job to the priority queue
func (p *NotificationWorkerPool) Submit(req *services.SendRequest) {
	p.jobQueue.mu.Lock()
	heap.Push(p.jobQueue, &Job{Request: req, Priority: req.Priority})
	p.jobQueue.mu.Unlock()
}

func (p *NotificationWorkerPool) worker(id int) {
	defer p.wg.Done()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.jobQueue.mu.Lock()
			if p.jobQueue.Len() > 0 {
				job := heap.Pop(p.jobQueue).(*Job)
				p.jobQueue.mu.Unlock()
				_ = p.service.SendNotification(p.ctx, job.Request)
			} else {
				p.jobQueue.mu.Unlock()
			}
		}
	}
}

// Stop gracefully stops the worker pool
func (p *NotificationWorkerPool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
}
