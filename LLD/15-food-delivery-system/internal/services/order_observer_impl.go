package services

import (
	"fmt"
	"food-delivery-system/internal/interfaces"
	"food-delivery-system/internal/models"
	"sync"
)

// OrderObserverManager manages multiple observers and notifies them (Observer Pattern)
type OrderObserverManager struct {
	observers []interfaces.OrderObserver
	mu        sync.RWMutex
}

// NewOrderObserverManager creates a new observer manager
func NewOrderObserverManager() *OrderObserverManager {
	return &OrderObserverManager{
		observers: []interfaces.OrderObserver{},
	}
}

// Subscribe adds an observer
func (m *OrderObserverManager) Subscribe(observer interfaces.OrderObserver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers = append(m.observers, observer)
}

// NotifyStatusChanged notifies all observers of status change
func (m *OrderObserverManager) NotifyStatusChanged(order *models.Order, oldStatus, newStatus models.OrderStatus) {
	m.mu.RLock()
	observers := make([]interfaces.OrderObserver, len(m.observers))
	copy(observers, m.observers)
	m.mu.RUnlock()

	for _, obs := range observers {
		obs.OnOrderStatusChanged(order, oldStatus, newStatus)
	}
}

// LoggingOrderObserver logs order status changes
type LoggingOrderObserver struct{}

func (l *LoggingOrderObserver) OnOrderStatusChanged(order *models.Order, oldStatus, newStatus models.OrderStatus) {
	fmt.Printf("[Order %s] Status: %s -> %s\n", order.ID, oldStatus, newStatus)
}
