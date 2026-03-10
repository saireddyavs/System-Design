package models

// CartItem represents a product line in the shopping cart
type CartItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	UnitPrice   float64 `json:"unit_price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// NewCartItem creates a cart item with calculated subtotal
func NewCartItem(productID, productName string, unitPrice float64, quantity int) CartItem {
	return CartItem{
		ProductID:   productID,
		ProductName: productName,
		UnitPrice:   unitPrice,
		Quantity:    quantity,
		Subtotal:    unitPrice * float64(quantity),
	}
}

// RecalculateSubtotal updates subtotal based on unit price and quantity
func (c *CartItem) RecalculateSubtotal() {
	c.Subtotal = c.UnitPrice * float64(c.Quantity)
}
