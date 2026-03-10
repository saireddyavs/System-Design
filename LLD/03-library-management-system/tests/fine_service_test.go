package tests

import (
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"library-management-system/internal/notifications"
	"library-management-system/internal/repositories"
	"library-management-system/internal/services"
	"testing"
	"time"
)

func setupFineTest(t *testing.T) (*services.FineService, interfaces.FineRepository, interfaces.LoanRepository, interfaces.MemberRepository, interfaces.BookRepository) {
	bookRepo := repositories.NewInMemoryBookRepo()
	memberRepo := repositories.NewInMemoryMemberRepo()
	loanRepo := repositories.NewInMemoryLoanRepo()
	fineRepo := repositories.NewInMemoryFineRepo()
	notifMgr := notifications.NewNotificationManager()
	notifMgr.Register(notifications.NewConsoleNotifier())
	fineCalc := services.NewPerDayFineCalculator()

	fineSvc := services.NewFineService(services.FineServiceConfig{
		FineRepo:       fineRepo,
		LoanRepo:       loanRepo,
		FineCalculator: fineCalc,
		Notifier:       notifMgr,
		MemberRepo:     memberRepo,
		BookRepo:       bookRepo,
	})

	return fineSvc, fineRepo, loanRepo, memberRepo, bookRepo
}

func TestProcessOverdueLoans_CreatesFine(t *testing.T) {
	fineSvc, fineRepo, loanRepo, memberRepo, bookRepo := setupFineTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	loanSvc := createLoanServiceForTest(bookRepo, memberRepo, loanRepo, fineRepo)

	book, _ := libSvc.AddBook("Overdue Book", "Author", "111", "Subject", 1)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)
	loan, _ := loanSvc.CheckOut(book.ID, member.ID)

	// Manually set loan to overdue (in real scenario, time would have passed)
	loan.DueDate = time.Now().Add(-5 * 24 * time.Hour) // 5 days ago
	_ = loanRepo.Update(loan)

	fines, err := fineSvc.ProcessOverdueLoans()
	if err != nil {
		t.Fatalf("ProcessOverdueLoans failed: %v", err)
	}
	if len(fines) != 1 {
		t.Fatalf("expected 1 fine, got %d", len(fines))
	}
	if fines[0].Amount != 5.0 {
		t.Errorf("expected fine $5 (5 days), got $%.2f", fines[0].Amount)
	}
	if fines[0].Status != models.FineStatusPending {
		t.Errorf("expected Pending status, got %s", fines[0].Status)
	}

	// Verify fine is in repo
	stored, _ := fineRepo.GetByLoanID(loan.ID)
	if stored == nil {
		t.Error("fine not found in repository")
	}
}

func TestPayFine_Success(t *testing.T) {
	fineSvc, fineRepo, _, _, _ := setupFineTest(t)

	fine := &models.Fine{
		ID:        "fine-1",
		LoanID:    "loan-1",
		MemberID:  "member-1",
		Amount:    10.0,
		Status:    models.FineStatusPending,
		CreatedAt: time.Now(),
	}
	_ = fineRepo.Create(fine)

	err := fineSvc.PayFine(fine.ID)
	if err != nil {
		t.Fatalf("PayFine failed: %v", err)
	}

	updated, _ := fineRepo.GetByID(fine.ID)
	if updated.Status != models.FineStatusPaid {
		t.Errorf("expected Paid status, got %s", updated.Status)
	}
	if updated.PaidAt == nil {
		t.Error("PaidAt should be set")
	}
}

func TestPayFine_AlreadyPaid(t *testing.T) {
	fineSvc, fineRepo, _, _, _ := setupFineTest(t)

	now := time.Now()
	fine := &models.Fine{
		ID:        "fine-2",
		LoanID:    "loan-2",
		MemberID:  "member-2",
		Amount:    5.0,
		Status:    models.FineStatusPaid,
		CreatedAt: now,
		PaidAt:    &now,
	}
	_ = fineRepo.Create(fine)

	err := fineSvc.PayFine(fine.ID)
	if err != services.ErrFineAlreadyPaid {
		t.Errorf("expected ErrFineAlreadyPaid, got %v", err)
	}
}

func TestGetMemberPendingFines(t *testing.T) {
	fineSvc, fineRepo, _, _, _ := setupFineTest(t)

	fine1 := &models.Fine{ID: "f1", LoanID: "l1", MemberID: "m1", Amount: 3, Status: models.FineStatusPending, CreatedAt: time.Now()}
	fine2 := &models.Fine{ID: "f2", LoanID: "l2", MemberID: "m1", Amount: 7, Status: models.FineStatusPending, CreatedAt: time.Now()}
	fine3 := &models.Fine{ID: "f3", LoanID: "l3", MemberID: "m1", Amount: 2, Status: models.FineStatusPaid, CreatedAt: time.Now()}
	_ = fineRepo.Create(fine1)
	_ = fineRepo.Create(fine2)
	_ = fineRepo.Create(fine3)

	fines, err := fineSvc.GetMemberPendingFines("m1")
	if err != nil {
		t.Fatalf("GetMemberPendingFines failed: %v", err)
	}
	if len(fines) != 2 {
		t.Errorf("expected 2 pending fines, got %d", len(fines))
	}
}

func createLoanServiceForTest(bookRepo interfaces.BookRepository, memberRepo interfaces.MemberRepository, loanRepo interfaces.LoanRepository, fineRepo interfaces.FineRepository) *services.LoanService {
	reservationRepo := repositories.NewInMemoryReservationRepo()
	notifMgr := notifications.NewNotificationManager()
	fineCalc := services.NewPerDayFineCalculator()

	return services.NewLoanService(services.LoanServiceConfig{
		BookRepo:        bookRepo,
		MemberRepo:      memberRepo,
		LoanRepo:        loanRepo,
		ReservationRepo: reservationRepo,
		FineRepo:        fineRepo,
		FineCalculator:  fineCalc,
		Notifier:        notifMgr,
	})
}
