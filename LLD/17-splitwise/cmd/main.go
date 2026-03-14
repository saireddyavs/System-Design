package main

import (
	"fmt"
	"log"

	"splitwise/internal/models"
	"splitwise/internal/repositories"
	"splitwise/internal/services"
	"splitwise/internal/strategies"
)

func main() {
	// Repositories
	userRepo := repositories.NewInMemoryUserRepository()
	groupRepo := repositories.NewInMemoryGroupRepository()
	expenseRepo := repositories.NewInMemoryExpenseRepository()
	balanceRepo := repositories.NewInMemoryBalanceRepository()
	transactionRepo := repositories.NewInMemoryTransactionRepository()

	// Split strategy registry (Factory Pattern)
	splitRegistry := strategies.NewSplitStrategyRegistry()

	// Services
	userService := services.NewUserService(userRepo)
	groupService := services.NewGroupService(groupRepo, userRepo)
	expenseService := services.NewExpenseService(expenseRepo, balanceRepo, groupRepo, userRepo, splitRegistry)
	balanceService := services.NewBalanceService(balanceRepo, transactionRepo)
	notificationService := services.NewNotificationService()

	// Register observers (Observer Pattern)
	expenseService.RegisterObserver(notificationService)
	balanceService.RegisterObserver(notificationService)

	fmt.Println("=== Splitwise Demo ===")

	// 1. Create 4 users
	fmt.Println("1. Creating 4 users...")
	users := make([]*models.User, 4)
	for i := 0; i < 4; i++ {
		u, err := userService.CreateUser(
			fmt.Sprintf("User%d", i+1),
			fmt.Sprintf("user%d@example.com", i+1),
			fmt.Sprintf("+123456789%d", i),
		)
		if err != nil {
			log.Fatalf("Create user failed: %v", err)
		}
		users[i] = u
		fmt.Printf("   Created: %s (%s)\n", u.Name, u.ID)
	}

	// 2. Create group "Trip to Goa"
	fmt.Println("\n2. Creating group 'Trip to Goa'...")
	group, err := groupService.CreateGroup("Trip to Goa", "Weekend trip to Goa", users[0].ID, []string{users[1].ID, users[2].ID, users[3].ID})
	if err != nil {
		log.Fatalf("Create group failed: %v", err)
	}
	fmt.Printf("   Created group: %s (ID: %s) with %d members\n", group.Name, group.ID, len(group.MemberIDs))

	// 3. Add expense: User1 paid $1000 for hotel, EQUAL among all 4
	fmt.Println("\n3. Adding expense: Hotel $1000 (EQUAL split)...")
	participants := []string{users[0].ID, users[1].ID, users[2].ID, users[3].ID}
	_, err = expenseService.AddExpense("Hotel", 1000, users[0].ID, models.SplitTypeEqual, participants, nil, group.ID)
	if err != nil {
		log.Fatalf("Add expense failed: %v", err)
	}
	fmt.Println("   Added: $1000 / 4 = $250 each")

	// 4. Add expense: User2 paid $600 for dinner, PERCENTAGE (40%, 30%, 20%, 10%)
	fmt.Println("\n4. Adding expense: Dinner $600 (PERCENTAGE split 40%, 30%, 20%, 10%)...")
	percentParams := map[string]float64{
		users[0].ID: 40,
		users[1].ID: 30,
		users[2].ID: 20,
		users[3].ID: 10,
	}
	_, err = expenseService.AddExpense("Dinner", 600, users[1].ID, models.SplitTypePercentage, participants, percentParams, group.ID)
	if err != nil {
		log.Fatalf("Add expense failed: %v", err)
	}
	fmt.Println("   Added: $600 split by percentage")

	// 5. Add expense: User3 paid $300 for taxi, EXACT amounts
	fmt.Println("\n5. Adding expense: Taxi $300 (EXACT split $100, $80, $70, $50)...")
	exactParams := map[string]float64{
		users[0].ID: 100,
		users[1].ID: 80,
		users[2].ID: 70,
		users[3].ID: 50,
	}
	_, err = expenseService.AddExpense("Taxi", 300, users[2].ID, models.SplitTypeExact, participants, exactParams, group.ID)
	if err != nil {
		log.Fatalf("Add expense failed: %v", err)
	}
	fmt.Println("   Added: $300 with exact amounts")

	// 6. Show all balances
	fmt.Println("\n6. Current balances (who owes whom):")
	balances, err := balanceService.GetBalancesForGroup(group.ID)
	if err != nil {
		log.Fatalf("Get balances failed: %v", err)
	}
	for _, b := range balances {
		debtor, _ := userService.GetUser(b.DebtorID)
		creditor, _ := userService.GetUser(b.CreditorID)
		debtorName := b.DebtorID
		creditorName := b.CreditorID
		if debtor != nil {
			debtorName = debtor.Name
		}
		if creditor != nil {
			creditorName = creditor.Name
		}
		fmt.Printf("   %s owes %s: $%.2f\n", debtorName, creditorName, b.Amount)
	}

	// 7. Simplify debts
	fmt.Println("\n7. Simplified debt settlement (minimize transactions):")
	simplified, err := balanceService.SimplifyDebts(group.ID)
	if err != nil {
		log.Fatalf("Simplify debts failed: %v", err)
	}
	for _, tx := range simplified {
		from, _ := userService.GetUser(tx.FromUserID)
		to, _ := userService.GetUser(tx.ToUserID)
		fromName := tx.FromUserID
		toName := tx.ToUserID
		if from != nil {
			fromName = from.Name
		}
		if to != nil {
			toName = to.Name
		}
		fmt.Printf("   %s pays %s: $%.2f\n", fromName, toName, tx.Amount)
	}

	// 8. Record a settlement (use an actual direct balance - simplified shows ideal flow)
	if len(balances) > 0 {
		fmt.Println("\n8. Recording settlement (direct balance)...")
		b := balances[0]
		settled, err := balanceService.Settle(b.DebtorID, b.CreditorID, b.Amount, group.ID)
		if err != nil {
			log.Fatalf("Settle failed: %v", err)
		}
		from, _ := userService.GetUser(settled.FromUserID)
		to, _ := userService.GetUser(settled.ToUserID)
		fmt.Printf("   Recorded: %s paid $%.2f to %s\n", from.Name, settled.Amount, to.Name)
	}

	// 9. Show updated balances
	fmt.Println("\n9. Updated balances after settlement:")
	balances, err = balanceService.GetBalancesForGroup(group.ID)
	if err != nil {
		log.Fatalf("Get balances failed: %v", err)
	}
	if len(balances) == 0 {
		fmt.Println("   All settled! No outstanding balances.")
	} else {
		for _, b := range balances {
			debtor, _ := userService.GetUser(b.DebtorID)
			creditor, _ := userService.GetUser(b.CreditorID)
			debtorName := b.DebtorID
			creditorName := b.CreditorID
			if debtor != nil {
				debtorName = debtor.Name
			}
			if creditor != nil {
				creditorName = creditor.Name
			}
			fmt.Printf("   %s owes %s: $%.2f\n", debtorName, creditorName, b.Amount)
		}
	}

	// Show notification log
	fmt.Println("\n10. Notification log:")
	for _, msg := range notificationService.GetExpenseLog() {
		fmt.Printf("   [Expense] %s\n", msg)
	}
	for _, msg := range notificationService.GetSettlementLog() {
		fmt.Printf("   [Settlement] %s\n", msg)
	}

	fmt.Println("\n=== Demo Complete ===")
}
