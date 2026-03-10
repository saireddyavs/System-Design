package tests

import (
	"atm-system/internal/models"
	"atm-system/internal/repositories"
	"atm-system/internal/services"
	"context"
	"testing"
	"time"
)

func setupAuthTest(t *testing.T) (*services.AuthService, *repositories.InMemoryAccountRepository) {
	accountRepo := repositories.NewInMemoryAccountRepository()
	cardRepo := accountRepo
	ctx := context.Background()

	acc := models.NewAccount("acc-1", "1234567890", "Test User", models.AccountTypeChecking, 10000, "1234")
	_ = accountRepo.Save(ctx, acc)
	card := models.NewCard("card-1", "4111111111111111", "acc-1", time.Now().Add(5*365*24*time.Hour))
	_ = cardRepo.SaveCard(ctx, card)

	return services.NewAuthService(accountRepo, cardRepo), accountRepo
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	authSvc, _ := setupAuthTest(t)
	ctx := context.Background()

	result, err := authSvc.Authenticate(ctx, "4111111111111111", "1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %s", result.Message)
	}
	if result.Account == nil {
		t.Fatal("expected account to be set")
	}
	if result.Account.HolderName != "Test User" {
		t.Errorf("expected holder Test User, got %s", result.Account.HolderName)
	}
}

func TestAuthService_Authenticate_WrongPIN(t *testing.T) {
	authSvc, _ := setupAuthTest(t)
	ctx := context.Background()

	result, _ := authSvc.Authenticate(ctx, "4111111111111111", "9999")
	if result.Success {
		t.Fatal("expected failure for wrong PIN")
	}
	if result.AttemptsLeft != 2 {
		t.Errorf("expected 2 attempts left, got %d", result.AttemptsLeft)
	}
}

func TestAuthService_Authenticate_CardBlockedAfter3Attempts(t *testing.T) {
	authSvc, accountRepo := setupAuthTest(t)
	ctx := context.Background()

	// 3 wrong attempts
	for i := 0; i < 3; i++ {
		authSvc.Authenticate(ctx, "4111111111111111", "9999")
	}

	result, _ := authSvc.Authenticate(ctx, "4111111111111111", "1234")
	if result.Success {
		t.Fatal("expected failure - card should be blocked")
	}
	if result.Message != "card blocked due to too many wrong PIN attempts" {
		t.Errorf("unexpected message: %s", result.Message)
	}

	// Verify card is actually blocked in repo
	card, _ := accountRepo.GetByCardNumber(ctx, "4111111111111111")
	if !card.IsBlocked {
		t.Fatal("card should be blocked in repository")
	}
}

func TestAuthService_ValidateCard_BlockedCard(t *testing.T) {
	authSvc, accountRepo := setupAuthTest(t)
	ctx := context.Background()

	card, _ := accountRepo.GetByCardNumber(ctx, "4111111111111111")
	card.IsBlocked = true
	_ = accountRepo.UpdateCard(ctx, card)

	_, err := authSvc.ValidateCard(ctx, "4111111111111111")
	if err == nil {
		t.Fatal("expected error for blocked card")
	}
}

func TestAuthService_ValidateCard_ExpiredCard(t *testing.T) {
	authSvc, accountRepo := setupAuthTest(t)
	ctx := context.Background()

	card, _ := accountRepo.GetByCardNumber(ctx, "4111111111111111")
	card.ExpiryDate = time.Now().Add(-24 * time.Hour)
	_ = accountRepo.UpdateCard(ctx, card)

	_, err := authSvc.ValidateCard(ctx, "4111111111111111")
	if err == nil {
		t.Fatal("expected error for expired card")
	}
}
