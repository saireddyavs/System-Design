package interfaces

// PaymentProcessor defines payment operations (Interface Segregation - ISP)
type PaymentProcessor interface {
	ProcessPayment(amount float64, bookingID string) error
	ProcessRefund(amount float64, bookingID string) error
}
