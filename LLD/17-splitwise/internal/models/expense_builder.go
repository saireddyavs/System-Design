package models

import (
	"crypto/rand"
	"fmt"
	"time"
)

// ExpenseBuilder builds Expense objects (Builder Pattern)
type ExpenseBuilder struct {
	expense *Expense
}

// NewExpenseBuilder creates a new expense builder
func NewExpenseBuilder() *ExpenseBuilder {
	b := make([]byte, 8)
	rand.Read(b)
	return &ExpenseBuilder{
		expense: &Expense{
			ID:        fmt.Sprintf("exp-%x", b),
			Status:    ExpenseStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// WithID sets the expense ID
func (b *ExpenseBuilder) WithID(id string) *ExpenseBuilder {
	b.expense.ID = id
	return b
}

// WithDescription sets the expense description
func (b *ExpenseBuilder) WithDescription(desc string) *ExpenseBuilder {
	b.expense.Description = desc
	return b
}

// WithAmount sets the expense amount
func (b *ExpenseBuilder) WithAmount(amount float64) *ExpenseBuilder {
	b.expense.Amount = amount
	return b
}

// WithPaidBy sets who paid the expense
func (b *ExpenseBuilder) WithPaidBy(userID string) *ExpenseBuilder {
	b.expense.PaidBy = userID
	return b
}

// WithSplitType sets the split type
func (b *ExpenseBuilder) WithSplitType(splitType SplitType) *ExpenseBuilder {
	b.expense.SplitType = splitType
	return b
}

// WithSplits sets the pre-calculated splits
func (b *ExpenseBuilder) WithSplits(splits []Split) *ExpenseBuilder {
	b.expense.Splits = splits
	return b
}

// WithGroupID sets the group (empty for non-group expense)
func (b *ExpenseBuilder) WithGroupID(groupID string) *ExpenseBuilder {
	b.expense.GroupID = groupID
	return b
}

// Build returns the built expense
func (b *ExpenseBuilder) Build() *Expense {
	b.expense.UpdatedAt = time.Now()
	return b.expense
}
