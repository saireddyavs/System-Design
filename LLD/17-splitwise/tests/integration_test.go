package tests

import (
	"testing"

	"splitwise/internal/models"
	"splitwise/internal/repositories"
	"splitwise/internal/services"
	"splitwise/internal/strategies"
)

func TestFullFlow(t *testing.T) {
	userRepo := repositories.NewInMemoryUserRepository()
	groupRepo := repositories.NewInMemoryGroupRepository()
	expenseRepo := repositories.NewInMemoryExpenseRepository()
	balanceRepo := repositories.NewInMemoryBalanceRepository()
	transactionRepo := repositories.NewInMemoryTransactionRepository()

	userService := services.NewUserService(userRepo)
	groupService := services.NewGroupService(groupRepo, userRepo)
	registry := strategies.NewSplitStrategyRegistry()
	expenseService := services.NewExpenseService(expenseRepo, balanceRepo, groupRepo, userRepo, registry)
	balanceService := services.NewBalanceService(balanceRepo, transactionRepo)

	// Create users
	u1, _ := userService.CreateUser("Alice", "a@x.com", "111")
	u2, _ := userService.CreateUser("Bob", "b@x.com", "222")
	u3, _ := userService.CreateUser("Charlie", "c@x.com", "333")

	// Create group
	group, _ := groupService.CreateGroup("Trip", "Weekend", u1.ID, []string{u2.ID, u3.ID})
	participants := []string{u1.ID, u2.ID, u3.ID}

	// Add equal expense
	_, err := expenseService.AddExpense("Hotel", 300, u1.ID, models.SplitTypeEqual, participants, nil, group.ID)
	if err != nil {
		t.Fatalf("expense 1 failed: %v", err)
	}

	// Add percentage expense
	params := map[string]float64{u1.ID: 50, u2.ID: 30, u3.ID: 20}
	_, err = expenseService.AddExpense("Dinner", 200, u2.ID, models.SplitTypePercentage, participants, params, group.ID)
	if err != nil {
		t.Fatalf("expense 2 failed: %v", err)
	}

	// Check balances
	balances, err := balanceService.GetBalancesForGroup(group.ID)
	if err != nil {
		t.Fatalf("get balances failed: %v", err)
	}
	if len(balances) == 0 {
		t.Fatal("expected some balances")
	}

	// Simplify and settle
	simplified, _ := balanceService.SimplifyDebts(group.ID)
	if len(simplified) > 0 {
		tx := simplified[0]
		_, err = balanceService.Settle(tx.FromUserID, tx.ToUserID, tx.Amount, group.ID)
		if err != nil {
			t.Fatalf("settle failed: %v", err)
		}
	}

	// Verify expense history
	expenses, _ := expenseService.GetExpensesByGroup(group.ID)
	if len(expenses) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(expenses))
	}
}
