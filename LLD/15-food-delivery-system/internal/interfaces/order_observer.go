package interfaces

import "food-delivery-system/internal/models"

// OrderObserver defines the contract for order status notifications (Observer Pattern)
type OrderObserver interface {
	OnOrderStatusChanged(order *models.Order, oldStatus, newStatus models.OrderStatus)
}
