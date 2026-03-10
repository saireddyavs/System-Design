package services

import (
	"context"
	"log"

	"ecommerce-website/internal/interfaces"
)

// LoggingNotificationService logs notifications (implements NotificationService for Observer pattern)
type LoggingNotificationService struct{}

// NewLoggingNotificationService creates a notification service that logs to stdout
func NewLoggingNotificationService() *LoggingNotificationService {
	return &LoggingNotificationService{}
}

// Notify implements interfaces.NotificationService
func (s *LoggingNotificationService) Notify(ctx context.Context, payload interfaces.NotificationPayload) error {
	switch payload.Type {
	case interfaces.NotificationLowStock:
		log.Printf("[LOW STOCK] Product %s (%s) has %d units left\n", payload.ProductName, payload.ProductID, payload.Stock)
	case interfaces.NotificationOrderStatus:
		log.Printf("[ORDER STATUS] Order %s for user %s: %s\n", payload.OrderID, payload.UserID, payload.Status)
	}
	return nil
}

var _ interfaces.NotificationService = (*LoggingNotificationService)(nil)
