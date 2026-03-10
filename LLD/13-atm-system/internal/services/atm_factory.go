package services

import (
	"atm-system/internal/hardware"
	"atm-system/internal/models"
	"atm-system/internal/repositories"
	"context"
	"sync"
	"time"
)

var (
	atmInstance *ATMService
	atmOnce     sync.Once
)

// GetATMInstance returns the singleton ATM instance (Singleton Pattern)
func GetATMInstance() *ATMService {
	atmOnce.Do(func() {
		atmInstance = createATMInstance()
	})
	return atmInstance
}

func createATMInstance() *ATMService {
	// Initialize cash inventory
	cashInventory := models.CashInventory{
		models.Denomination100:  50,
		models.Denomination500:  40,
		models.Denomination1000: 30,
		models.Denomination2000: 20,
	}

	atm := models.NewATM("ATM-001", "Main Branch", cashInventory)

	accountRepo := repositories.NewInMemoryAccountRepository()
	cardRepo := accountRepo // Same struct implements both
	txRepo := repositories.NewInMemoryTransactionRepository()

	// Seed test data
	seedTestData(accountRepo, cardRepo)

	authService := NewAuthService(accountRepo, cardRepo)
	cashDispenser := hardware.NewGreedyCashDispenser()
	receiptPrinter := hardware.NewReceiptPrinter()
	validator := BuildValidationChain()
	accountSvc := NewAccountService(accountRepo)
	transactionSvc := NewTransactionService(txRepo)

	return NewATMService(
		atm,
		authService,
		cashDispenser,
		receiptPrinter,
		accountRepo,
		validator,
		accountSvc,
		transactionSvc,
	)
}

func seedTestData(accountRepo *repositories.InMemoryAccountRepository, cardRepo *repositories.InMemoryAccountRepository) {
	ctx := context.Background()

	// Account 1: 100000 balance, PIN 1234
	acc1 := models.NewAccount("acc-1", "1234567890", "John Doe", models.AccountTypeChecking, 100000, "1234")
	_ = accountRepo.Save(ctx, acc1)
	card1 := models.NewCard("card-1", "4111111111111111", "acc-1", time.Now().Add(5*365*24*time.Hour))
	_ = cardRepo.SaveCard(ctx, card1)

	// Account 2: 50000 balance, PIN 5678
	acc2 := models.NewAccount("acc-2", "0987654321", "Jane Smith", models.AccountTypeSavings, 50000, "5678")
	_ = accountRepo.Save(ctx, acc2)
	card2 := models.NewCard("card-2", "4222222222222222", "acc-2", time.Now().Add(5*365*24*time.Hour))
	_ = cardRepo.SaveCard(ctx, card2)
}
