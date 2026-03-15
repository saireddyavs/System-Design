package hardware

import (
	"atm-system/internal/interfaces"
	"context"
	"fmt"
	"strings"
)

// ReceiptPrinter implements receipt printing (simulated hardware)
type ReceiptPrinter struct{}

// NewReceiptPrinter creates a new receipt printer
func NewReceiptPrinter() *ReceiptPrinter {
	return &ReceiptPrinter{}
}

var _ interfaces.ReceiptPrinter = (*ReceiptPrinter)(nil)

// Print generates a receipt string (simulates printing)
func (p *ReceiptPrinter) Print(ctx context.Context, data *interfaces.ReceiptData) (string, error) {
	var sb strings.Builder

	sb.WriteString("================================\n")
	sb.WriteString("         ATM RECEIPT\n")
	sb.WriteString("================================\n")
	sb.WriteString(fmt.Sprintf("Date/Time: %s\n", data.Timestamp))
	sb.WriteString(fmt.Sprintf("Account:   %s\n", maskAccountNumber(data.AccountNumber)))
	sb.WriteString(fmt.Sprintf("Ref ID:    %s\n", data.ReferenceID))
	sb.WriteString("--------------------------------\n")
	sb.WriteString(fmt.Sprintf("Transaction: %s\n", data.TransactionType))
	if data.Amount > 0 {
		sb.WriteString(fmt.Sprintf("Amount:     Rs. %.2f\n", data.Amount))
	}
	sb.WriteString(fmt.Sprintf("Balance:    Rs. %.2f\n", data.Balance))
	sb.WriteString("--------------------------------\n")
	sb.WriteString("Thank you for banking with us!\n")
	sb.WriteString("================================\n")

	return sb.String(), nil
}

// maskAccountNumber masks middle digits for security
func maskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 4 {
		return "****"
	}
	if len(accountNumber) <= 8 {
		return accountNumber[:2] + "****" + accountNumber[len(accountNumber)-2:]
	}
	return accountNumber[:4] + strings.Repeat("*", len(accountNumber)-8) + accountNumber[len(accountNumber)-4:]
}
