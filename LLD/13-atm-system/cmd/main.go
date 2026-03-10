package main

import (
	"atm-system/internal/services"
	"context"
	"fmt"
	"log"
)

func main() {
	// Get singleton ATM instance
	atm := services.GetATMInstance()
	ctx := context.Background()

	fmt.Println("=== ATM System Demo ===")
	fmt.Println()

	// 1. Insert card
	fmt.Println("1. Inserting card 4111111111111111...")
	if err := atm.InsertCard(ctx, "4111111111111111"); err != nil {
		log.Fatalf("Insert card failed: %v", err)
	}
	fmt.Printf("   ATM State: %s\n\n", atm.GetState())

	// 2. Authenticate
	fmt.Println("2. Authenticating with PIN 1234...")
	result, err := atm.Authenticate(ctx, "4111111111111111", "1234")
	if err != nil {
		log.Fatalf("Auth failed: %v", err)
	}
	if !result.Success {
		log.Fatalf("Auth failed: %s", result.Message)
	}
	fmt.Printf("   Authenticated as %s\n\n", result.Account.HolderName)

	// 3. Balance inquiry
	fmt.Println("3. Balance inquiry...")
	balanceResult, err := atm.GetBalance(ctx)
	if err != nil {
		log.Fatalf("Balance inquiry failed: %v", err)
	}
	fmt.Printf("   %s\n\n", balanceResult.Message)

	// 4. Withdraw
	fmt.Println("4. Withdrawing Rs. 2500...")
	withdrawResult, err := atm.Withdraw(ctx, 2500)
	if err != nil {
		log.Fatalf("Withdraw failed: %v", err)
	}
	fmt.Printf("   %s\n", withdrawResult.Message)
	if withdrawResult.Success && withdrawResult.Data != nil {
		fmt.Printf("   Denominations: %v\n", withdrawResult.Data)
	}

	// 5. Print receipt
	receipt, _ := atm.PrintReceipt(ctx, withdrawResult)
	if receipt != "" {
		fmt.Println("\n   Receipt:")
		fmt.Println(receipt)
	}

	// 6. Mini statement
	fmt.Println("\n5. Mini statement (last 5)...")
	stmtResult, err := atm.GetMiniStatement(ctx, 5)
	if err != nil {
		log.Fatalf("Mini statement failed: %v", err)
	}
	if stmtResult.Success && stmtResult.Data != nil {
		fmt.Printf("   %s\n", stmtResult.Message)
	}

	// 7. Eject card
	fmt.Println("\n6. Ejecting card...")
	if err := atm.EjectCard(ctx); err != nil {
		log.Fatalf("Eject failed: %v", err)
	}
	fmt.Printf("   ATM State: %s\n", atm.GetState())

	fmt.Println("\n=== Demo Complete ===")
}
