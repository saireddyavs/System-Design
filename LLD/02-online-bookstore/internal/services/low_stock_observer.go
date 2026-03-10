package services

import (
	"log"

	"online-bookstore/internal/models"
)

// LowStockObserver implements InventoryObserver for low-stock notifications.
// Observer pattern: Can be replaced with email/Slack/alerting in production.
type LowStockObserver struct{}

func NewLowStockObserver() *LowStockObserver {
	return &LowStockObserver{}
}

func (o *LowStockObserver) OnLowStock(book *models.Book, threshold int) {
	log.Printf("[LOW STOCK ALERT] Book '%s' (ID: %s) has %d units left (threshold: %d). Consider restocking.",
		book.Title, book.ID, book.Stock, threshold)
}
