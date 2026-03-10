package services

import (
	"atm-system/internal/interfaces"
	"atm-system/internal/models"
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidState      = errors.New("invalid ATM state for this operation")
	ErrATMMaintenance    = errors.New("ATM is out of service")
	ErrSessionExpired    = errors.New("session expired")
	ErrNotAuthenticated  = errors.New("user not authenticated")
)

// ATMService orchestrates ATM operations with State Pattern
type ATMService struct {
	atm              *models.ATM
	authService      interfaces.AuthService
	cashDispenser    interfaces.CashDispenser
	receiptPrinter   interfaces.ReceiptPrinter
	accountRepo      interfaces.AccountRepository
	validator        interfaces.TransactionValidator
	accountSvc       *AccountService
	transactionSvc   *TransactionService
	currentSession   *Session
	mu               sync.RWMutex
}

// Session holds authenticated user session
type Session struct {
	AccountID   string
	Account     *models.Account
	CardNumber  string
	AuthenticatedAt time.Time
}

// NewATMService creates a new ATM service (Singleton-like - one ATM instance)
func NewATMService(
	atm *models.ATM,
	authService interfaces.AuthService,
	cashDispenser interfaces.CashDispenser,
	receiptPrinter interfaces.ReceiptPrinter,
	accountRepo interfaces.AccountRepository,
	validator interfaces.TransactionValidator,
	accountSvc *AccountService,
	transactionSvc *TransactionService,
) *ATMService {
	return &ATMService{
		atm:            atm,
		authService:    authService,
		cashDispenser:  cashDispenser,
		receiptPrinter: receiptPrinter,
		accountRepo:    accountRepo,
		validator:      validator,
		accountSvc:     accountSvc,
		transactionSvc: transactionSvc,
	}
}

// InsertCard - State: Idle -> CardInserted
func (s *ATMService) InsertCard(ctx context.Context, cardNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.atm.GetState() != models.ATMStateIdle {
		return ErrInvalidState
	}
	if s.atm.IsOutOfService() {
		s.atm.SetState(models.ATMStateOutOfService)
		return ErrATMMaintenance
	}

	card, err := s.authService.ValidateCard(ctx, cardNumber)
	if err != nil {
		return err
	}

	s.atm.SetState(models.ATMStateCardInserted)
	s.atm.CurrentCardID = card.ID
	return nil
}

// Authenticate - State: CardInserted -> Authenticated
func (s *ATMService) Authenticate(ctx context.Context, cardNumber, pin string) (*interfaces.AuthResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.atm.GetState() != models.ATMStateCardInserted {
		return &interfaces.AuthResult{Success: false, Message: ErrInvalidState.Error()}, nil
	}

	result, err := s.authService.Authenticate(ctx, cardNumber, pin)
	if err != nil {
		return result, err
	}
	if !result.Success {
		return result, nil
	}

	s.atm.SetState(models.ATMStateAuthenticated)
	s.currentSession = &Session{
		AccountID:       result.Account.ID,
		Account:         result.Account,
		CardNumber:      cardNumber,
		AuthenticatedAt: time.Now(),
	}
	return result, nil
}

// ExecuteCommand - State: Authenticated -> TransactionInProgress, then back to Authenticated
func (s *ATMService) ExecuteCommand(ctx context.Context, cmd interfaces.ATMCommand) (*interfaces.CommandResult, error) {
	s.mu.Lock()
	if s.atm.GetState() != models.ATMStateAuthenticated && s.atm.GetState() != models.ATMStateTransactionInProgress {
		s.mu.Unlock()
		return &interfaces.CommandResult{Success: false, Message: ErrInvalidState.Error()}, nil
	}
	if s.currentSession == nil {
		s.mu.Unlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	if time.Since(s.currentSession.AuthenticatedAt) > models.TransactionTimeoutSecs*time.Second {
		s.mu.Unlock()
		s.ejectCard()
		return &interfaces.CommandResult{Success: false, Message: ErrSessionExpired.Error()}, nil
	}

	// Refresh account from repo (for latest balance)
	account, err := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	if err != nil {
		s.mu.Unlock()
		return &interfaces.CommandResult{Success: false, Message: err.Error()}, err
	}
	s.currentSession.Account = account

	s.atm.SetState(models.ATMStateTransactionInProgress)
	s.mu.Unlock()

	// Execute command (may need to pass cmdCtx - commands already have context from constructor)
	// For commands that need context, we need to refactor - they get ctx in constructor
	result, err := cmd.Execute(ctx)

	s.mu.Lock()
	s.atm.SetState(models.ATMStateAuthenticated)
	if s.atm.IsOutOfService() {
		s.atm.SetState(models.ATMStateOutOfService)
		s.ejectCard()
	}
	s.mu.Unlock()

	return result, err
}

// EjectCard - State: Authenticated/CardInserted -> Idle
func (s *ATMService) EjectCard(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ejectCard()
}

func (s *ATMService) ejectCard() error {
	state := s.atm.GetState()
	if state != models.ATMStateCardInserted && state != models.ATMStateAuthenticated && state != models.ATMStateOutOfService {
		return ErrInvalidState
	}
	s.atm.SetState(models.ATMStateIdle)
	s.atm.CurrentCardID = ""
	s.atm.CurrentPIN = ""
	s.currentSession = nil
	return nil
}

// GetBalance executes balance inquiry
func (s *ATMService) GetBalance(ctx context.Context) (*interfaces.CommandResult, error) {
	s.mu.RLock()
	if s.currentSession == nil {
		s.mu.RUnlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	account, _ := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	cmdCtx := &CommandContext{AccountID: s.currentSession.AccountID, Account: account}
	s.mu.RUnlock()

	cmd := NewBalanceInquiryCommand(cmdCtx, s.accountSvc, s.transactionSvc)
	return s.ExecuteCommand(ctx, cmd)
}

// Withdraw executes withdrawal
func (s *ATMService) Withdraw(ctx context.Context, amount float64) (*interfaces.CommandResult, error) {
	s.mu.RLock()
	if s.currentSession == nil {
		s.mu.RUnlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	account, _ := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	cmdCtx := &CommandContext{AccountID: s.currentSession.AccountID, Account: account}
	s.mu.RUnlock()

	cmd := NewWithdrawalCommand(cmdCtx, amount, s.validator, s.cashDispenser, s.accountSvc, s.transactionSvc, s.atm)
	return s.ExecuteCommand(ctx, cmd)
}

// Deposit executes deposit
func (s *ATMService) Deposit(ctx context.Context, amount float64) (*interfaces.CommandResult, error) {
	s.mu.RLock()
	if s.currentSession == nil {
		s.mu.RUnlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	account, _ := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	cmdCtx := &CommandContext{AccountID: s.currentSession.AccountID, Account: account}
	s.mu.RUnlock()

	cmd := NewDepositCommand(cmdCtx, amount, s.accountSvc, s.transactionSvc)
	return s.ExecuteCommand(ctx, cmd)
}

// ChangePIN executes PIN change
func (s *ATMService) ChangePIN(ctx context.Context, newPIN string) (*interfaces.CommandResult, error) {
	s.mu.RLock()
	if s.currentSession == nil {
		s.mu.RUnlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	account, _ := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	cmdCtx := &CommandContext{AccountID: s.currentSession.AccountID, Account: account}
	s.mu.RUnlock()

	cmd := NewPINChangeCommand(cmdCtx, newPIN, s.accountSvc, s.transactionSvc)
	return s.ExecuteCommand(ctx, cmd)
}

// GetMiniStatement executes mini statement
func (s *ATMService) GetMiniStatement(ctx context.Context, limit int) (*interfaces.CommandResult, error) {
	s.mu.RLock()
	if s.currentSession == nil {
		s.mu.RUnlock()
		return &interfaces.CommandResult{Success: false, Message: ErrNotAuthenticated.Error()}, nil
	}
	account, _ := s.accountRepo.GetByID(ctx, s.currentSession.AccountID)
	cmdCtx := &CommandContext{AccountID: s.currentSession.AccountID, Account: account}
	s.mu.RUnlock()

	cmd := NewMiniStatementCommand(cmdCtx, limit, s.transactionSvc)
	return s.ExecuteCommand(ctx, cmd)
}

// PrintReceipt prints a receipt for the last transaction
func (s *ATMService) PrintReceipt(ctx context.Context, result *interfaces.CommandResult) (string, error) {
	if result == nil || !result.Success || result.Transaction == nil {
		return "", nil
	}
	s.mu.RLock()
	accountNumber := ""
	if s.currentSession != nil && s.currentSession.Account != nil {
		accountNumber = s.currentSession.Account.AccountNumber
	}
	s.mu.RUnlock()

	data := &interfaces.ReceiptData{
		TransactionType: string(result.Transaction.Type),
		Amount:          result.Transaction.Amount,
		Balance:         result.Transaction.BalanceAfter,
		Timestamp:       result.Transaction.Timestamp.Format("2006-01-02 15:04:05"),
		AccountNumber:   accountNumber,
		ReferenceID:     result.Transaction.ID,
	}
	return s.receiptPrinter.Print(ctx, data)
}

// GetState returns current ATM state
func (s *ATMService) GetState() models.ATMState {
	return s.atm.GetState()
}
