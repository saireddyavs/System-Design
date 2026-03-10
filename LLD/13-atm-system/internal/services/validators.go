package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
	"fmt"
)

// BaseValidator provides common chain of responsibility logic
type BaseValidator struct {
	next interfaces.TransactionValidator
}

func (v *BaseValidator) SetNext(validator interfaces.TransactionValidator) {
	v.next = validator
}

// BalanceValidator validates sufficient balance (Chain of Responsibility)
type BalanceValidator struct {
	BaseValidator
}

func NewBalanceValidator() *BalanceValidator {
	return &BalanceValidator{}
}

func (v *BalanceValidator) Validate(ctx context.Context, account *models.Account, amount float64, txType models.TransactionType) *interfaces.ValidationResult {
	if txType != models.TransactionTypeWithdrawal {
		if v.next != nil {
			return v.next.Validate(ctx, account, amount, txType)
		}
		return &interfaces.ValidationResult{Valid: true}
	}
	if account.GetBalance() < amount {
		return &interfaces.ValidationResult{
			Valid:   false,
			Message: "insufficient balance",
		}
	}
	if v.next != nil {
		return v.next.Validate(ctx, account, amount, txType)
	}
	return &interfaces.ValidationResult{Valid: true}
}

// DailyLimitValidator validates daily withdrawal limit
type DailyLimitValidator struct {
	BaseValidator
}

func NewDailyLimitValidator() *DailyLimitValidator {
	return &DailyLimitValidator{}
}

func (v *DailyLimitValidator) Validate(ctx context.Context, account *models.Account, amount float64, txType models.TransactionType) *interfaces.ValidationResult {
	if txType != models.TransactionTypeWithdrawal {
		if v.next != nil {
			return v.next.Validate(ctx, account, amount, txType)
		}
		return &interfaces.ValidationResult{Valid: true}
	}
	dailyWithdrawn := account.GetDailyWithdrawn()
	if dailyWithdrawn+amount > models.DailyWithdrawalLimit {
		return &interfaces.ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("daily withdrawal limit exceeded (limit: %d, withdrawn: %.0f)", models.DailyWithdrawalLimit, dailyWithdrawn),
		}
	}
	if v.next != nil {
		return v.next.Validate(ctx, account, amount, txType)
	}
	return &interfaces.ValidationResult{Valid: true}
}

// AmountValidator validates withdrawal amount rules
type AmountValidator struct {
	BaseValidator
}

func NewAmountValidator() *AmountValidator {
	return &AmountValidator{}
}

func (v *AmountValidator) Validate(ctx context.Context, account *models.Account, amount float64, txType models.TransactionType) *interfaces.ValidationResult {
	if txType != models.TransactionTypeWithdrawal {
		if v.next != nil {
			return v.next.Validate(ctx, account, amount, txType)
		}
		return &interfaces.ValidationResult{Valid: true}
	}
	if amount < models.MinWithdrawalAmount {
		return &interfaces.ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("minimum withdrawal amount is %d", models.MinWithdrawalAmount),
		}
	}
	if int(amount)%models.MinWithdrawalAmount != 0 {
		return &interfaces.ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("amount must be multiple of %d", models.MinWithdrawalAmount),
		}
	}
	if v.next != nil {
		return v.next.Validate(ctx, account, amount, txType)
	}
	return &interfaces.ValidationResult{Valid: true}
}

// BuildValidationChain creates the complete validation chain
func BuildValidationChain() interfaces.TransactionValidator {
	amountValidator := NewAmountValidator()
	balanceValidator := NewBalanceValidator()
	dailyLimitValidator := NewDailyLimitValidator()

	amountValidator.SetNext(balanceValidator)
	balanceValidator.SetNext(dailyLimitValidator)

	return amountValidator
}
