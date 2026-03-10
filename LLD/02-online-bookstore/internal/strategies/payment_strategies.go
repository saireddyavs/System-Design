package strategies

import (
	"fmt"
	"time"

	"online-bookstore/internal/interfaces"
	"online-bookstore/internal/models"
)

// CreditCardProcessor implements PaymentProcessor for credit card payments.
// Strategy pattern: encapsulates credit card payment logic.
type CreditCardProcessor struct{}

func NewCreditCardProcessor() *CreditCardProcessor {
	return &CreditCardProcessor{}
}

func (p *CreditCardProcessor) Process(payment *models.Payment) (*models.Payment, error) {
	// Simulate payment processing - in production, integrate with payment gateway
	payment.TransactionID = fmt.Sprintf("CC-%d", time.Now().UnixNano())
	payment.Status = models.PaymentStatusCompleted
	return payment, nil
}

func (p *CreditCardProcessor) GetMethodName() string {
	return "credit_card"
}

// PayPalProcessor implements PaymentProcessor for PayPal payments.
type PayPalProcessor struct{}

func NewPayPalProcessor() *PayPalProcessor {
	return &PayPalProcessor{}
}

func (p *PayPalProcessor) Process(payment *models.Payment) (*models.Payment, error) {
	payment.TransactionID = fmt.Sprintf("PP-%d", time.Now().UnixNano())
	payment.Status = models.PaymentStatusCompleted
	return payment, nil
}

func (p *PayPalProcessor) GetMethodName() string {
	return "paypal"
}

// PaymentProcessorRegistry holds available payment processors (Strategy selection).
// OCP: Add new payment methods by registering new processors.
type PaymentProcessorRegistry struct {
	processors map[string]interfaces.PaymentProcessor
}

func NewPaymentProcessorRegistry() *PaymentProcessorRegistry {
	registry := &PaymentProcessorRegistry{
		processors: make(map[string]interfaces.PaymentProcessor),
	}
	// Register default processors
	registry.Register(NewCreditCardProcessor())
	registry.Register(NewPayPalProcessor())
	return registry
}

func (r *PaymentProcessorRegistry) Register(processor interfaces.PaymentProcessor) {
	r.processors[processor.GetMethodName()] = processor
}

func (r *PaymentProcessorRegistry) GetProcessor(method string) (interfaces.PaymentProcessor, bool) {
	p, ok := r.processors[method]
	return p, ok
}
