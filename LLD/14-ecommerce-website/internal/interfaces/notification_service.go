package interfaces

import (
	"context"
	"ecommerce-website/internal/models"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationLowStock    NotificationType = "LowStock"
	NotificationOrderStatus NotificationType = "OrderStatus"
)

// NotificationPayload represents the data sent in a notification
type NotificationPayload struct {
	Type       NotificationType
	ProductID  string
	ProductName string
	Stock      int
	OrderID    string
	Status     models.OrderStatus
	UserID     string
}

// NotificationService defines the interface for sending notifications (Observer pattern)
type NotificationService interface {
	Notify(ctx context.Context, payload NotificationPayload) error
}
