package strategies

import (
	"context"
	"fmt"
	"sync/atomic"

	"ecommerce-website/internal/interfaces"
	"ecommerce-website/internal/models"
)

// CreditCardProcessor processes credit card payments
type CreditCardProcessor struct{}

func NewCreditCardProcessor() *CreditCardProcessor {
	return &CreditCardProcessor{}
}

func (p *CreditCardProcessor) Process(ctx context.Context, payment *models.Payment) error {
	// Simulate payment processing
	payment.Status = models.PaymentStatusCompleted
	payment.TransactionID = fmt.Sprintf("CC-%d", atomic.AddUint64(&txnCounter, 1))
	return nil
}

func (p *CreditCardProcessor) GetMethod() models.PaymentMethod {
	return models.PaymentMethodCreditCard
}

// DebitCardProcessor processes debit card payments
type DebitCardProcessor struct{}

func NewDebitCardProcessor() *DebitCardProcessor {
	return &DebitCardProcessor{}
}

func (p *DebitCardProcessor) Process(ctx context.Context, payment *models.Payment) error {
	payment.Status = models.PaymentStatusCompleted
	payment.TransactionID = fmt.Sprintf("DC-%d", atomic.AddUint64(&txnCounter, 1))
	return nil
}

func (p *DebitCardProcessor) GetMethod() models.PaymentMethod {
	return models.PaymentMethodDebitCard
}

// UPIProcessor processes UPI payments
type UPIProcessor struct{}

func NewUPIProcessor() *UPIProcessor {
	return &UPIProcessor{}
}

func (p *UPIProcessor) Process(ctx context.Context, payment *models.Payment) error {
	payment.Status = models.PaymentStatusCompleted
	payment.TransactionID = fmt.Sprintf("UPI-%d", atomic.AddUint64(&txnCounter, 1))
	return nil
}

func (p *UPIProcessor) GetMethod() models.PaymentMethod {
	return models.PaymentMethodUPI
}

// WalletProcessor processes wallet payments
type WalletProcessor struct{}

func NewWalletProcessor() *WalletProcessor {
	return &WalletProcessor{}
}

func (p *WalletProcessor) Process(ctx context.Context, payment *models.Payment) error {
	payment.Status = models.PaymentStatusCompleted
	payment.TransactionID = fmt.Sprintf("WALLET-%d", atomic.AddUint64(&txnCounter, 1))
	return nil
}

func (p *WalletProcessor) GetMethod() models.PaymentMethod {
	return models.PaymentMethodWallet
}

var txnCounter uint64

// Ensure all processors implement the interface
var _ interfaces.PaymentProcessor = (*CreditCardProcessor)(nil)
var _ interfaces.PaymentProcessor = (*DebitCardProcessor)(nil)
var _ interfaces.PaymentProcessor = (*UPIProcessor)(nil)
var _ interfaces.PaymentProcessor = (*WalletProcessor)(nil)
