package interfaces

import "shopping-cart-system/internal/models"

// CartEventType represents cart lifecycle events
type CartEventType string

const (
	CartEventItemAdded    CartEventType = "item_added"
	CartEventItemRemoved  CartEventType = "item_removed"
	CartEventItemUpdated  CartEventType = "item_updated"
	CartEventCheckout     CartEventType = "checkout"
	CartEventAbandoned    CartEventType = "abandoned"
)

// CartEventType fix - I made a typo above
type CartEvent struct {
	Type   CartEventType
	Cart   *models.Cart
	UserID string
}

// CartEventObserver defines observer for cart events (Observer pattern)
type CartEventObserver interface {
	OnCartEvent(event CartEvent)
}
