package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
	"fmt"
	"time"
)

// CommandContext holds context for command execution
type CommandContext struct {
	AccountID string
	Account   *models.Account
}

// BalanceInquiryCommand implements balance inquiry (Command Pattern)
type BalanceInquiryCommand struct {
	ctx         *CommandContext
	accountSvc  *AccountService
	transactionSvc *TransactionService
}

func NewBalanceInquiryCommand(ctx *CommandContext, accountSvc *AccountService, txSvc *TransactionService) *BalanceInquiryCommand {
	return &BalanceInquiryCommand{ctx: ctx, accountSvc: accountSvc, transactionSvc: txSvc}
}

func (c *BalanceInquiryCommand) Execute(ctx context.Context) (*interfaces.CommandResult, error) {
	balance := c.ctx.Account.GetBalance()
	tx, err := c.transactionSvc.CreateTransaction(ctx, c.ctx.AccountID, models.TransactionTypeBalanceInquiry, 0, balance, balance)
	if err != nil {
		return &interfaces.CommandResult{Success: false, Message: err.Error()}, err
	}
	return &interfaces.CommandResult{
		Success:     true,
		Transaction: tx,
		Data:        balance,
		Message:     fmt.Sprintf("Balance: Rs. %.2f", balance),
	}, nil
}

func (c *BalanceInquiryCommand) GetType() models.TransactionType {
	return models.TransactionTypeBalanceInquiry
}

// WithdrawalCommand implements cash withdrawal (Command Pattern)
type WithdrawalCommand struct {
	ctx            *CommandContext
	amount         float64
	validator      interfaces.TransactionValidator
	cashDispenser  interfaces.CashDispenser
	accountSvc     *AccountService
	transactionSvc *TransactionService
	atm            *models.ATM
}

func NewWithdrawalCommand(ctx *CommandContext, amount float64, validator interfaces.TransactionValidator, dispenser interfaces.CashDispenser, accountSvc *AccountService, txSvc *TransactionService, atm *models.ATM) *WithdrawalCommand {
	return &WithdrawalCommand{
		ctx:            ctx,
		amount:         amount,
		validator:      validator,
		cashDispenser:  dispenser,
		accountSvc:     accountSvc,
		transactionSvc: txSvc,
		atm:            atm,
	}
}

func (c *WithdrawalCommand) Execute(ctx context.Context) (*interfaces.CommandResult, error) {
	// Chain of Responsibility: Validate
	if result := c.validator.Validate(ctx, c.ctx.Account, c.amount, models.TransactionTypeWithdrawal); !result.Valid {
		return &interfaces.CommandResult{Success: false, Message: result.Message}, nil
	}

	// Check ATM cash
	inventory := c.atm.GetCashInventory()
	if !c.cashDispenser.CanDispense(ctx, c.amount, inventory) {
		return &interfaces.CommandResult{Success: false, Message: "ATM cannot dispense requested amount with available denominations"}, nil
	}

	balanceBefore := c.ctx.Account.GetBalance()

	// Deduct from account
	if !c.ctx.Account.DeductBalance(c.amount) {
		return &interfaces.CommandResult{Success: false, Message: "insufficient balance"}, nil
	}

	// Record withdrawal for daily limit
	c.ctx.Account.RecordWithdrawal(c.amount)

	// Dispense cash (Strategy: Greedy)
	dispenseResult, err := c.cashDispenser.Dispense(ctx, c.amount, inventory)
	if err != nil || !dispenseResult.Success {
		// Rollback
		c.ctx.Account.AddBalance(c.amount)
		c.ctx.Account.RecordWithdrawal(-c.amount)
		msg := "dispense failed"
		if dispenseResult != nil {
			msg = dispenseResult.ErrorMessage
		}
		return &interfaces.CommandResult{Success: false, Message: msg}, err
	}

	// Update ATM inventory
	c.atm.UpdateCashInventory(dispenseResult.Dispensed)

	balanceAfter := c.ctx.Account.GetBalance()
	tx, _ := c.transactionSvc.CreateTransaction(ctx, c.ctx.AccountID, models.TransactionTypeWithdrawal, c.amount, balanceBefore, balanceAfter)

	return &interfaces.CommandResult{
		Success:     true,
		Transaction: tx,
		Data:        dispenseResult.Dispensed,
		Message:     fmt.Sprintf("Withdrawn Rs. %.2f successfully", c.amount),
	}, nil
}

func (c *WithdrawalCommand) GetType() models.TransactionType {
	return models.TransactionTypeWithdrawal
}

// DepositCommand implements cash deposit (Command Pattern)
type DepositCommand struct {
	ctx            *CommandContext
	amount         float64
	accountSvc     *AccountService
	transactionSvc *TransactionService
}

func NewDepositCommand(ctx *CommandContext, amount float64, accountSvc *AccountService, txSvc *TransactionService) *DepositCommand {
	return &DepositCommand{ctx: ctx, amount: amount, accountSvc: accountSvc, transactionSvc: txSvc}
}

func (c *DepositCommand) Execute(ctx context.Context) (*interfaces.CommandResult, error) {
	if c.amount <= 0 {
		return &interfaces.CommandResult{Success: false, Message: "invalid deposit amount"}, nil
	}
	balanceBefore := c.ctx.Account.GetBalance()
	c.ctx.Account.AddBalance(c.amount)
	balanceAfter := c.ctx.Account.GetBalance()

	tx, err := c.transactionSvc.CreateTransaction(ctx, c.ctx.AccountID, models.TransactionTypeDeposit, c.amount, balanceBefore, balanceAfter)
	if err != nil {
		c.ctx.Account.DeductBalance(c.amount) // Rollback
		return &interfaces.CommandResult{Success: false, Message: err.Error()}, err
	}

	return &interfaces.CommandResult{
		Success:     true,
		Transaction: tx,
		Data:        balanceAfter,
		Message:     fmt.Sprintf("Deposited Rs. %.2f successfully", c.amount),
	}, nil
}

func (c *DepositCommand) GetType() models.TransactionType {
	return models.TransactionTypeDeposit
}

// PINChangeCommand implements PIN change (Command Pattern)
type PINChangeCommand struct {
	ctx         *CommandContext
	newPIN      string
	accountSvc  *AccountService
	transactionSvc *TransactionService
}

func NewPINChangeCommand(ctx *CommandContext, newPIN string, accountSvc *AccountService, txSvc *TransactionService) *PINChangeCommand {
	return &PINChangeCommand{ctx: ctx, newPIN: newPIN, accountSvc: accountSvc, transactionSvc: txSvc}
}

func (c *PINChangeCommand) Execute(ctx context.Context) (*interfaces.CommandResult, error) {
	if len(c.newPIN) != 4 {
		return &interfaces.CommandResult{Success: false, Message: "PIN must be 4 digits"}, nil
	}
	balance := c.ctx.Account.GetBalance()
	if err := c.accountSvc.UpdatePIN(ctx, c.ctx.AccountID, c.newPIN); err != nil {
		return &interfaces.CommandResult{Success: false, Message: err.Error()}, err
	}
	tx, _ := c.transactionSvc.CreateTransaction(ctx, c.ctx.AccountID, models.TransactionTypePINChange, 0, balance, balance)
	return &interfaces.CommandResult{
		Success:     true,
		Transaction: tx,
		Message:     "PIN changed successfully",
	}, nil
}

func (c *PINChangeCommand) GetType() models.TransactionType {
	return models.TransactionTypePINChange
}

// MiniStatementCommand implements mini statement (Command Pattern)
type MiniStatementCommand struct {
	ctx            *CommandContext
	limit          int
	transactionSvc *TransactionService
}

func NewMiniStatementCommand(ctx *CommandContext, limit int, txSvc *TransactionService) *MiniStatementCommand {
	if limit <= 0 {
		limit = 10
	}
	return &MiniStatementCommand{ctx: ctx, limit: limit, transactionSvc: txSvc}
}

func (c *MiniStatementCommand) Execute(ctx context.Context) (*interfaces.CommandResult, error) {
	transactions, err := c.transactionSvc.GetMiniStatement(ctx, c.ctx.AccountID, c.limit)
	if err != nil {
		return &interfaces.CommandResult{Success: false, Message: err.Error()}, err
	}
	return &interfaces.CommandResult{
		Success: true,
		Data:    transactions,
		Message: fmt.Sprintf("Last %d transactions", len(transactions)),
	}, nil
}

func (c *MiniStatementCommand) GetType() models.TransactionType {
	return models.TransactionTypeMiniStatement
}

// FormatMiniStatement formats transactions for display
func FormatMiniStatement(transactions []*models.Transaction) string {
	var result string
	for i, tx := range transactions {
		result += fmt.Sprintf("%d. %s | %s | Rs.%.2f | %s\n",
			i+1, tx.Type, tx.Timestamp.Format(time.RFC3339), tx.Amount, tx.Status)
	}
	return result
}
