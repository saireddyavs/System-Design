package tests

import (
	"testing"

	"splitwise/internal/models"
	"splitwise/internal/repositories"
	"splitwise/internal/services"
	"splitwise/internal/strategies"
)

func setupExpenseTest(t *testing.T) (*services.ExpenseService, *repositories.InMemoryGroupRepository, []*models.User) {
	userRepo := repositories.NewInMemoryUserRepository()
	groupRepo := repositories.NewInMemoryGroupRepository()
	expenseRepo := repositories.NewInMemoryExpenseRepository()
	balanceRepo := repositories.NewInMemoryBalanceRepository()

	// Create users
	users := []*models.User{
		{ID: "u1", Name: "Alice", Email: "a@x.com", Phone: "111"},
		{ID: "u2", Name: "Bob", Email: "b@x.com", Phone: "222"},
		{ID: "u3", Name: "Charlie", Email: "c@x.com", Phone: "333"},
	}
	for _, u := range users {
		_ = userRepo.Create(u)
	}

	// Create group
	group := &models.Group{
		ID:        "g1",
		Name:      "Test Group",
		MemberIDs: []string{"u1", "u2", "u3"},
		CreatedBy: "u1",
	}
	_ = groupRepo.Create(group)

	registry := strategies.NewSplitStrategyRegistry()
	expenseService := services.NewExpenseService(expenseRepo, balanceRepo, groupRepo, userRepo, registry)
	return expenseService, groupRepo, users
}

func TestAddEqualExpense(t *testing.T) {
	expenseService, _, users := setupExpenseTest(t)
	participants := []string{users[0].ID, users[1].ID, users[2].ID}

	exp, err := expenseService.AddExpense("Lunch", 300, users[0].ID, models.SplitTypeEqual, participants, nil, "g1")
	if err != nil {
		t.Fatalf("add expense failed: %v", err)
	}
	if exp.Amount != 300 {
		t.Errorf("amount should be 300, got %.2f", exp.Amount)
	}
	if len(exp.Splits) != 3 {
		t.Errorf("expected 3 splits, got %d", len(exp.Splits))
	}
	for _, s := range exp.Splits {
		if s.Amount != 100 {
			t.Errorf("each split should be 100, got %.2f", s.Amount)
		}
	}
}

func TestAddPercentageExpense(t *testing.T) {
	expenseService, _, users := setupExpenseTest(t)
	participants := []string{users[0].ID, users[1].ID, users[2].ID}
	params := map[string]float64{"u1": 50, "u2": 30, "u3": 20}

	exp, err := expenseService.AddExpense("Dinner", 1000, users[0].ID, models.SplitTypePercentage, participants, params, "g1")
	if err != nil {
		t.Fatalf("add expense failed: %v", err)
	}
	if exp.Amount != 1000 {
		t.Errorf("amount should be 1000, got %.2f", exp.Amount)
	}
	if len(exp.Splits) != 3 {
		t.Errorf("expected 3 splits, got %d", len(exp.Splits))
	}
}

func TestGetExpensesByGroup(t *testing.T) {
	expenseService, _, users := setupExpenseTest(t)
	participants := []string{users[0].ID, users[1].ID, users[2].ID}

	_, _ = expenseService.AddExpense("Exp1", 100, users[0].ID, models.SplitTypeEqual, participants, nil, "g1")
	_, _ = expenseService.AddExpense("Exp2", 200, users[1].ID, models.SplitTypeEqual, participants, nil, "g1")

	expenses, err := expenseService.GetExpensesByGroup("g1")
	if err != nil {
		t.Fatalf("get expenses failed: %v", err)
	}
	if len(expenses) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(expenses))
	}
}
