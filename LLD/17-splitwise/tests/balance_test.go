package tests

import (
	"math"
	"testing"

	"splitwise/internal/repositories"
	"splitwise/internal/services"
)

func setupBalanceTest(t *testing.T) (*services.BalanceService, *repositories.InMemoryBalanceRepository, string) {
	balanceRepo := repositories.NewInMemoryBalanceRepository()
	transactionRepo := repositories.NewInMemoryTransactionRepository()
	balanceService := services.NewBalanceService(balanceRepo, transactionRepo)

	// Add some balances: u1 owes u2 $50, u2 owes u3 $30, u1 owes u3 $20
	_ = balanceRepo.AddBalance("u1", "u2", "g1", 50)
	_ = balanceRepo.AddBalance("u2", "u3", "g1", 30)
	_ = balanceRepo.AddBalance("u1", "u3", "g1", 20)

	return balanceService, balanceRepo, "g1"
}

func TestSimplifyDebts(t *testing.T) {
	balanceService, _, groupID := setupBalanceTest(t)

	// Net: u1=-70, u2=+20, u3=+50
	// Simplified: u1 pays u3 $50, u1 pays u2 $20
	simplified, err := balanceService.SimplifyDebts(groupID)
	if err != nil {
		t.Fatalf("simplify failed: %v", err)
	}
	if len(simplified) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(simplified))
	}
	var totalAmount float64
	for _, tx := range simplified {
		if tx.FromUserID != "u1" {
			t.Errorf("debtor should be u1, got %s", tx.FromUserID)
		}
		totalAmount += tx.Amount
	}
	if math.Abs(totalAmount-70) > 0.01 {
		t.Errorf("total settlement should be 70, got %.2f", totalAmount)
	}
}

func TestSettle(t *testing.T) {
	balanceRepo := repositories.NewInMemoryBalanceRepository()
	transactionRepo := repositories.NewInMemoryTransactionRepository()
	balanceService := services.NewBalanceService(balanceRepo, transactionRepo)

	_ = balanceRepo.AddBalance("u1", "u2", "g1", 100)

	tx, err := balanceService.Settle("u1", "u2", 50, "g1")
	if err != nil {
		t.Fatalf("settle failed: %v", err)
	}
	if tx.Amount != 50 {
		t.Errorf("amount should be 50, got %.2f", tx.Amount)
	}

	// Remaining balance should be 50
	balances, _ := balanceService.GetBalancesForGroup("g1")
	if len(balances) != 1 {
		t.Fatalf("expected 1 balance, got %d", len(balances))
	}
	if balances[0].Amount != 50 {
		t.Errorf("remaining balance should be 50, got %.2f", balances[0].Amount)
	}
}

func TestSettleInsufficientBalance(t *testing.T) {
	balanceRepo := repositories.NewInMemoryBalanceRepository()
	transactionRepo := repositories.NewInMemoryTransactionRepository()
	balanceService := services.NewBalanceService(balanceRepo, transactionRepo)

	_ = balanceRepo.AddBalance("u1", "u2", "g1", 50)

	_, err := balanceService.Settle("u1", "u2", 100, "g1")
	if err == nil {
		t.Fatal("expected error for over-settlement")
	}
}
