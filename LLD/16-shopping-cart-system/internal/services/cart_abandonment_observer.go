package services

import (
	"log"

	"shopping-cart-system/internal/interfaces"
)

// CartAbandonmentObserver logs/tracks cart events (Observer pattern implementation)
type CartAbandonmentObserver struct {
	cartRepo interfaces.CartRepository
}

func NewCartAbandonmentObserver(cartRepo interfaces.CartRepository) *CartAbandonmentObserver {
	return &CartAbandonmentObserver{cartRepo: cartRepo}
}

func (o *CartAbandonmentObserver) OnCartEvent(event interfaces.CartEvent) {
	switch event.Type {
	case interfaces.CartEventItemAdded:
		log.Printf("[Observer] Item added to cart for user %s, cart has %d items", event.UserID, len(event.Cart.Items))
	case interfaces.CartEventItemRemoved:
		log.Printf("[Observer] Item removed from cart for user %s", event.UserID)
	case interfaces.CartEventItemUpdated:
		log.Printf("[Observer] Cart updated for user %s", event.UserID)
	case interfaces.CartEventCheckout:
		log.Printf("[Observer] Checkout completed for user %s, order total: %.2f", event.UserID, event.Cart.Subtotal())
	case interfaces.CartEventAbandoned:
		log.Printf("[Observer] Cart abandoned by user %s", event.UserID)
	}
}
