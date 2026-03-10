package models

// CartStatus represents the lifecycle state of a shopping cart
type CartStatus string

const (
	CartStatusActive     CartStatus = "active"
	CartStatusAbandoned  CartStatus = "abandoned"
	CartStatusCheckedOut CartStatus = "checked_out"
)

// CouponType defines discount calculation strategy
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage"
	CouponTypeFlat       CouponType = "flat"
	CouponTypeBOGO       CouponType = "bogo"
)

// OrderStatus represents order lifecycle
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// PaymentMethod represents payment strategy
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodWallet     PaymentMethod = "wallet"
)
