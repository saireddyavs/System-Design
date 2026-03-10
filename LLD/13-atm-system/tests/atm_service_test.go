package tests

import (
	"atm-system/internal/hardware"
	"atm-system/internal/models"
	"atm-system/internal/repositories"
	"atm-system/internal/services"
	"context"
	"sync"
	"testing"
	"time"
)

func setupATMService(t *testing.T) *services.ATMService {
	cashInventory := models.CashInventory{
		models.Denomination100:  50,
		models.Denomination500:  40,
		models.Denomination1000: 30,
		models.Denomination2000: 20,
	}
	atm := models.NewATM("ATM-TEST", "Test Branch", cashInventory)

	accountRepo := repositories.NewInMemoryAccountRepository()
	txRepo := repositories.NewInMemoryTransactionRepository()
	ctx := context.Background()

	acc := models.NewAccount("acc-1", "1234567890", "Test User", models.AccountTypeChecking, 50000, "1234")
	_ = accountRepo.Save(ctx, acc)
	card := models.NewCard("card-1", "4111111111111111", "acc-1", time.Now().Add(5*365*24*time.Hour))
	_ = accountRepo.SaveCard(ctx, card)

	authSvc := services.NewAuthService(accountRepo, accountRepo)
	dispenser := hardware.NewGreedyCashDispenser()
	receiptPrinter := hardware.NewReceiptPrinter()
	validator := services.BuildValidationChain()
	accountSvc := services.NewAccountService(accountRepo)
	transactionSvc := services.NewTransactionService(txRepo)

	return services.NewATMService(
		atm, authSvc, dispenser, receiptPrinter,
		accountRepo, validator, accountSvc, transactionSvc,
	)
}

func TestATMService_InsertCard_InvalidState(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	// Insert card first
	_ = atmSvc.InsertCard(ctx, "4111111111111111")
	// Try insert again - should fail
	err := atmSvc.InsertCard(ctx, "4111111111111111")
	if err == nil {
		t.Fatal("expected error when inserting card in non-idle state")
	}
}

func TestATMService_Withdraw_Success(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	_ = atmSvc.InsertCard(ctx, "4111111111111111")
	_, _ = atmSvc.Authenticate(ctx, "4111111111111111", "1234")

	result, err := atmSvc.Withdraw(ctx, 2500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Message)
	}
}

func TestATMService_Withdraw_ExceedsDailyLimit(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	_ = atmSvc.InsertCard(ctx, "4111111111111111")
	_, _ = atmSvc.Authenticate(ctx, "4111111111111111", "1234")

	// Try to withdraw more than daily limit (50000)
	result, _ := atmSvc.Withdraw(ctx, 51000)
	if result.Success {
		t.Fatal("expected failure for exceeding daily limit")
	}
}

func TestATMService_Withdraw_InsufficientBalance(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	_ = atmSvc.InsertCard(ctx, "4111111111111111")
	_, _ = atmSvc.Authenticate(ctx, "4111111111111111", "1234")

	result, _ := atmSvc.Withdraw(ctx, 100000)
	if result.Success {
		t.Fatal("expected failure for insufficient balance")
	}
}

func TestATMService_Withdraw_NotAuthenticated(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	// Don't authenticate
	_ = atmSvc.InsertCard(ctx, "4111111111111111")

	result, _ := atmSvc.Withdraw(ctx, 1000)
	if result.Success {
		t.Fatal("expected failure when not authenticated")
	}
}

func TestATMService_ConcurrentAccess(t *testing.T) {
	// Create two ATMs with shared account (simulates concurrent access to different ATMs)
	accountRepo := repositories.NewInMemoryAccountRepository()
	txRepo := repositories.NewInMemoryTransactionRepository()
	ctx := context.Background()

	acc := models.NewAccount("acc-1", "1234567890", "Test User", models.AccountTypeChecking, 100000, "1234")
	_ = accountRepo.Save(ctx, acc)
	card := models.NewCard("card-1", "4111111111111111", "acc-1", time.Now().Add(5*365*24*time.Hour))
	_ = accountRepo.SaveCard(ctx, card)

	cashInv := models.CashInventory{
		models.Denomination100:  100,
		models.Denomination500:  100,
		models.Denomination1000: 100,
		models.Denomination2000: 100,
	}

	createATM := func() *services.ATMService {
		atm := models.NewATM("ATM-1", "Branch", cashInv)
		authSvc := services.NewAuthService(accountRepo, accountRepo)
		return services.NewATMService(
			atm, authSvc, hardware.NewGreedyCashDispenser(),
			hardware.NewReceiptPrinter(), accountRepo,
			services.BuildValidationChain(),
			services.NewAccountService(accountRepo),
			services.NewTransactionService(txRepo),
		)
	}

	atm1 := createATM()
	atm2 := createATM()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = atm1.InsertCard(ctx, "4111111111111111")
		_, _ = atm1.Authenticate(ctx, "4111111111111111", "1234")
		_, _ = atm1.Withdraw(ctx, 1000)
		_ = atm1.EjectCard(ctx)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond) // Slight delay
		_ = atm2.InsertCard(ctx, "4111111111111111")
		_, _ = atm2.Authenticate(ctx, "4111111111111111", "1234")
		_, _ = atm2.Withdraw(ctx, 2000)
		_ = atm2.EjectCard(ctx)
	}()

	wg.Wait()

	// Verify final balance (100000 - 1000 - 2000 = 97000)
	account, _ := accountRepo.GetByID(ctx, "acc-1")
	balance := account.GetBalance()
	if balance != 97000 {
		t.Errorf("expected balance 97000, got %.2f", balance)
	}
}

func TestATMService_ValidationChain(t *testing.T) {
	atmSvc := setupATMService(t)
	ctx := context.Background()

	_ = atmSvc.InsertCard(ctx, "4111111111111111")
	_, _ = atmSvc.Authenticate(ctx, "4111111111111111", "1234")

	// Invalid amount - not multiple of 100
	result, _ := atmSvc.Withdraw(ctx, 250)
	if result.Success {
		t.Fatal("expected failure for amount 250 (not multiple of 100)")
	}

	// Invalid amount - below minimum
	result, _ = atmSvc.Withdraw(ctx, 50)
	if result.Success {
		t.Fatal("expected failure for amount below minimum")
	}
}
