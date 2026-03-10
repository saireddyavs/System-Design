package services

import (
	"ride-sharing-service/internal/interfaces"
	"ride-sharing-service/internal/models"
	"sync"
)

// RideNotifier implements Observer pattern - notifies observers of ride status changes
type RideNotifier struct {
	observers []interfaces.RideObserver
	mu        sync.RWMutex
}

// NewRideNotifier creates a new ride notifier
func NewRideNotifier() *RideNotifier {
	return &RideNotifier{
		observers: make([]interfaces.RideObserver, 0),
	}
}

// Subscribe adds an observer
func (n *RideNotifier) Subscribe(observer interfaces.RideObserver) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.observers = append(n.observers, observer)
}

// Notify notifies all observers of ride status change
func (n *RideNotifier) Notify(ride *models.Ride) {
	n.mu.RLock()
	observers := make([]interfaces.RideObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, obs := range observers {
		obs.OnRideStatusChanged(ride)
	}
}
