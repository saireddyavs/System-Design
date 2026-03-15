package broker

import (
	"chat-application/internal/interfaces"
	"chat-application/internal/models"
	"sync"
)

// DeliveryStrategy defines how messages are delivered (Strategy pattern)
type DeliveryStrategy interface {
	Deliver(broker *InMemoryBroker, message *models.Message, recipientIDs []string) error
}

// DirectDelivery delivers to specific recipients one-by-one
type DirectDelivery struct{}

func (d *DirectDelivery) Deliver(broker *InMemoryBroker, message *models.Message, recipientIDs []string) error {
	for _, userID := range recipientIDs {
		broker.deliverToUser(userID, message)
	}
	return nil
}

// InMemoryBroker implements MessageBroker - Observer pattern for real-time delivery
// Each user has a channel; subscribers receive messages via their channel
type InMemoryBroker struct {
	subscribers map[string]chan *models.Message
	queues      map[string][]*models.Message
	mu          sync.RWMutex
	strategy    DeliveryStrategy
}

// NewInMemoryBroker creates a new in-memory message broker
func NewInMemoryBroker(strategy DeliveryStrategy) interfaces.MessageBroker {
	if strategy == nil {
		strategy = &DirectDelivery{}
	}
	return &InMemoryBroker{
		subscribers: make(map[string]chan *models.Message),
		queues:      make(map[string][]*models.Message),
		strategy:    strategy,
	}
}

// Subscribe adds a subscriber - returns receive-only channel (Observer pattern)
func (b *InMemoryBroker) Subscribe(userID string) <-chan *models.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, exists := b.subscribers[userID]; exists {
		return ch
	}
	ch := make(chan *models.Message, 100)
	b.subscribers[userID] = ch
	return ch
}

// Unsubscribe removes a subscriber
func (b *InMemoryBroker) Unsubscribe(userID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, exists := b.subscribers[userID]; exists {
		close(ch)
		delete(b.subscribers, userID)
	}
}

// Publish delivers message to recipients using configured strategy
func (b *InMemoryBroker) Publish(message *models.Message, recipientIDs []string) error {
	// No lock here - deliverToUser acquires its own lock to avoid deadlock with QueueForOffline
	return b.strategy.Deliver(b, message, recipientIDs)
}

// deliverToUser sends message to user's channel (internal)
func (b *InMemoryBroker) deliverToUser(userID string, message *models.Message) {
	b.mu.Lock()
	ch, exists := b.subscribers[userID]
	b.mu.Unlock()

	if exists {
		select {
		case ch <- message:
			return
		default:
			// Channel full - queue for later
		}
	}
	b.QueueForOffline(userID, message)
}

// QueueForOffline queues message for offline users
func (b *InMemoryBroker) QueueForOffline(userID string, message *models.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	msgCopy := *message
	b.queues[userID] = append(b.queues[userID], &msgCopy)
}

// GetQueuedMessages returns queued messages for a user
func (b *InMemoryBroker) GetQueuedMessages(userID string) []*models.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	msgs := b.queues[userID]
	if len(msgs) == 0 {
		return nil
	}
	result := make([]*models.Message, len(msgs))
	copy(result, msgs)
	return result
}

// ClearQueue removes queued messages after delivery
func (b *InMemoryBroker) ClearQueue(userID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.queues, userID)
}
