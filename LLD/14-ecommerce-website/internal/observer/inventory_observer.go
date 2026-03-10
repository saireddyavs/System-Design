package observer

import (
	"context"
	"sync"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// InventoryObserver implements Observer pattern for low stock alerts
type InventoryObserver struct {
	mu              sync.RWMutex
	notifiers       []interfaces.NotificationService
	lowStockThreshold int
}

// NewInventoryObserver creates a new inventory observer
func NewInventoryObserver(threshold int) *InventoryObserver {
	return &InventoryObserver{
		notifiers:        make([]interfaces.NotificationService, 0),
		lowStockThreshold: threshold,
	}
}

// Subscribe adds a notification service as observer
func (o *InventoryObserver) Subscribe(notifier interfaces.NotificationService) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.notifiers = append(o.notifiers, notifier)
}

// NotifyLowStock checks stock and notifies if below threshold
func (o *InventoryObserver) NotifyLowStock(ctx context.Context, productID, productName string, stock int) {
	o.mu.RLock()
	threshold := o.lowStockThreshold
	notifiers := make([]interfaces.NotificationService, len(o.notifiers))
	copy(notifiers, o.notifiers)
	o.mu.RUnlock()

	if stock >= threshold {
		return
	}

	payload := interfaces.NotificationPayload{
		Type:       interfaces.NotificationLowStock,
		ProductID:  productID,
		ProductName: productName,
		Stock:      stock,
	}

	for _, n := range notifiers {
		_ = n.Notify(ctx, payload)
	}
}

// OrderStatusObserver notifies on order status changes
type OrderStatusObserver struct {
	mu        sync.RWMutex
	notifiers []interfaces.NotificationService
}

// NewOrderStatusObserver creates observer for order status
func NewOrderStatusObserver() *OrderStatusObserver {
	return &OrderStatusObserver{
		notifiers: make([]interfaces.NotificationService, 0),
	}
}

// Subscribe adds a notification service
func (o *OrderStatusObserver) Subscribe(notifier interfaces.NotificationService) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.notifiers = append(o.notifiers, notifier)
}

// NotifyOrderStatus notifies all observers of order status change
func (o *OrderStatusObserver) NotifyOrderStatus(ctx context.Context, orderID, userID string, status models.OrderStatus) {
	o.mu.RLock()
	notifiers := make([]interfaces.NotificationService, len(o.notifiers))
	copy(notifiers, o.notifiers)
	o.mu.RUnlock()

	payload := interfaces.NotificationPayload{
		Type:    interfaces.NotificationOrderStatus,
		OrderID: orderID,
		UserID:  userID,
		Status:  status,
	}

	for _, n := range notifiers {
		_ = n.Notify(ctx, payload)
	}
}
