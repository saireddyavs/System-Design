package strategies

import (
	"fmt"
	"shopping-cart-system/internal/interfaces"
	"shopping-cart-system/internal/models"
)

// CreditCardProcessor processes credit card payments
type CreditCardProcessor struct{}

func (p *CreditCardProcessor) Supports(method models.PaymentMethod) bool {
	return method == models.PaymentMethodCreditCard
}

func (p *CreditCardProcessor) Process(req *interfaces.PaymentRequest) (*interfaces.PaymentResult, error) {
	// Simulated payment - in production would integrate with payment gateway
	paymentID := fmt.Sprintf("cc_%s_%d", req.OrderID, len(req.OrderID))
	return &interfaces.PaymentResult{
		Success:   true,
		PaymentID: paymentID,
		Message:   "Payment successful via Credit Card",
	}, nil
}

// PayPalProcessor processes PayPal payments
type PayPalProcessor struct{}

func (p *PayPalProcessor) Supports(method models.PaymentMethod) bool {
	return method == models.PaymentMethodPayPal
}

func (p *PayPalProcessor) Process(req *interfaces.PaymentRequest) (*interfaces.PaymentResult, error) {
	paymentID := fmt.Sprintf("pp_%s_%d", req.OrderID, len(req.OrderID))
	return &interfaces.PaymentResult{
		Success:   true,
		PaymentID: paymentID,
		Message:   "Payment successful via PayPal",
	}, nil
}

// WalletProcessor processes wallet/digital wallet payments
type WalletProcessor struct{}

func (p *WalletProcessor) Supports(method models.PaymentMethod) bool {
	return method == models.PaymentMethodWallet
}

func (p *WalletProcessor) Process(req *interfaces.PaymentRequest) (*interfaces.PaymentResult, error) {
	paymentID := fmt.Sprintf("wallet_%s_%d", req.OrderID, len(req.OrderID))
	return &interfaces.PaymentResult{
		Success:   true,
		PaymentID: paymentID,
		Message:   "Payment successful via Wallet",
	}, nil
}

// PaymentProcessorRegistry holds all processors and selects by method
type PaymentProcessorRegistry struct {
	processors []interfaces.PaymentProcessor
}

func NewPaymentProcessorRegistry() *PaymentProcessorRegistry {
	return &PaymentProcessorRegistry{
		processors: []interfaces.PaymentProcessor{
			&CreditCardProcessor{},
			&PayPalProcessor{},
			&WalletProcessor{},
		},
	}
}

func (r *PaymentProcessorRegistry) GetProcessor(method models.PaymentMethod) (interfaces.PaymentProcessor, error) {
	for _, p := range r.processors {
		if p.Supports(method) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unsupported payment method: %s", method)
}
