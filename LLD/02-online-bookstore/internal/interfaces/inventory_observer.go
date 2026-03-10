package interfaces

import "online-bookstore/internal/models"

// InventoryObserver defines the contract for low-stock notifications.
// Observer pattern: Decouples inventory from notification mechanisms.
type InventoryObserver interface {
	OnLowStock(book *models.Book, threshold int)
}
