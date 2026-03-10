package interfaces

import "context"

// ReceiptData contains data to print on receipt
type ReceiptData struct {
	TransactionType string
	Amount          float64
	Balance         float64
	Timestamp       string
	AccountNumber   string
	ReferenceID     string
}

// ReceiptPrinter defines the interface for receipt printing
type ReceiptPrinter interface {
	Print(ctx context.Context, data *ReceiptData) (string, error)
}
