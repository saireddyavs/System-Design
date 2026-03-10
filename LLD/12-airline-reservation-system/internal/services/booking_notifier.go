package services

import (
	"airline-reservation-system/internal/interfaces"
	"airline-reservation-system/internal/models"
	"sync"
)

// BookingNotifier manages observers and dispatches booking events (Observer Pattern)
type BookingNotifier struct {
	observers []interfaces.BookingObserver
	mu        sync.RWMutex
}

// NewBookingNotifier creates a new booking notifier
func NewBookingNotifier() *BookingNotifier {
	return &BookingNotifier{
		observers: make([]interfaces.BookingObserver, 0),
	}
}

// Subscribe adds an observer
func (n *BookingNotifier) Subscribe(observer interfaces.BookingObserver) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.observers = append(n.observers, observer)
}

// NotifyBookingCreated notifies all observers of a new booking
func (n *BookingNotifier) NotifyBookingCreated(booking *models.Booking) {
	n.mu.RLock()
	observers := make([]interfaces.BookingObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, o := range observers {
		o.OnBookingCreated(booking)
	}
}

// NotifyBookingCancelled notifies all observers of a cancelled booking
func (n *BookingNotifier) NotifyBookingCancelled(booking *models.Booking) {
	n.mu.RLock()
	observers := make([]interfaces.BookingObserver, len(n.observers))
	copy(observers, n.observers)
	n.mu.RUnlock()

	for _, o := range observers {
		o.OnBookingCancelled(booking)
	}
}
