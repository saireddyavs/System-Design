package models

// SplitType defines how an expense is divided among participants
type SplitType string

const (
	SplitTypeEqual      SplitType = "EQUAL"
	SplitTypeExact      SplitType = "EXACT"
	SplitTypePercentage SplitType = "PERCENTAGE"
	SplitTypeShare      SplitType = "SHARE"
)

// ExpenseStatus represents the lifecycle of an expense
type ExpenseStatus string

const (
	ExpenseStatusActive   ExpenseStatus = "ACTIVE"
	ExpenseStatusSettled  ExpenseStatus = "SETTLED"
	ExpenseStatusDeleted  ExpenseStatus = "DELETED"
)
