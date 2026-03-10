package interfaces

// PaymentProcessor defines the contract for payment operations (Interface Segregation)
type PaymentProcessor interface {
	// ProcessPayment processes a payment and returns transaction ID or error
	ProcessPayment(amount float64, currency string, reference string) (string, error)
	// ProcessRefund processes a refund for a transaction
	ProcessRefund(transactionID string, amount float64) error
}
