package tests

import (
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"library-management-system/internal/notifications"
	"library-management-system/internal/repositories"
	"library-management-system/internal/services"
	"testing"
)

func setupReservationTest(t *testing.T) (*services.ReservationService, *services.LoanService, interfaces.BookRepository, interfaces.MemberRepository) {
	bookRepo := repositories.NewInMemoryBookRepo()
	memberRepo := repositories.NewInMemoryMemberRepo()
	loanRepo := repositories.NewInMemoryLoanRepo()
	reservationRepo := repositories.NewInMemoryReservationRepo()
	fineRepo := repositories.NewInMemoryFineRepo()
	notifMgr := notifications.NewNotificationManager()
	fineCalc := services.NewPerDayFineCalculator()

	loanSvc := services.NewLoanService(services.LoanServiceConfig{
		BookRepo:        bookRepo,
		MemberRepo:      memberRepo,
		LoanRepo:        loanRepo,
		ReservationRepo: reservationRepo,
		FineRepo:        fineRepo,
		FineCalculator:  fineCalc,
		Notifier:        notifMgr,
	})

	reservationSvc := services.NewReservationService(bookRepo, memberRepo, loanRepo, reservationRepo)
	return reservationSvc, loanSvc, bookRepo, memberRepo
}

func TestReserve_Success(t *testing.T) {
	reservationSvc, loanSvc, bookRepo, memberRepo := setupReservationTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)

	book, _ := libSvc.AddBook("Popular Book", "Author", "222", "Subject", 1)
	member1, _ := libSvc.RegisterMember("Alice", "alice@test.com", models.MembershipStandard)
	member2, _ := libSvc.RegisterMember("Bob", "bob@test.com", models.MembershipStandard)

	// Alice checks out the only copy
	_, _ = loanSvc.CheckOut(book.ID, member1.ID)

	// Bob reserves
	res, err := reservationSvc.Reserve(book.ID, member2.ID)
	if err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}
	if res.Status != models.ReservationStatusPending {
		t.Errorf("expected Pending status, got %s", res.Status)
	}
	if res.BookID != book.ID || res.MemberID != member2.ID {
		t.Error("reservation has wrong book or member ID")
	}
}

func TestReserve_BookAvailable(t *testing.T) {
	reservationSvc, _, bookRepo, memberRepo := setupReservationTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)

	book, _ := libSvc.AddBook("Available Book", "Author", "333", "Subject", 2)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	_, err := reservationSvc.Reserve(book.ID, member.ID)
	if err != services.ErrBookAvailable {
		t.Errorf("expected ErrBookAvailable, got %v", err)
	}
}

func TestReserve_DuplicateReservation(t *testing.T) {
	reservationSvc, loanSvc, bookRepo, memberRepo := setupReservationTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)

	book, _ := libSvc.AddBook("Book", "Author", "444", "Subject", 1)
	member1, _ := libSvc.RegisterMember("Alice", "alice@test.com", models.MembershipStandard)
	member2, _ := libSvc.RegisterMember("Bob", "bob@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member1.ID)
	_, _ = reservationSvc.Reserve(book.ID, member2.ID)
	_, err := reservationSvc.Reserve(book.ID, member2.ID)
	if err != services.ErrMemberAlreadyReserved {
		t.Errorf("expected ErrMemberAlreadyReserved, got %v", err)
	}
}

func TestCancelReservation_Success(t *testing.T) {
	reservationSvc, loanSvc, bookRepo, memberRepo := setupReservationTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)

	book, _ := libSvc.AddBook("Book", "Author", "555", "Subject", 1)
	member1, _ := libSvc.RegisterMember("Alice", "alice@test.com", models.MembershipStandard)
	member2, _ := libSvc.RegisterMember("Bob", "bob@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member1.ID)
	res, _ := reservationSvc.Reserve(book.ID, member2.ID)

	err := reservationSvc.CancelReservation(res.ID, member2.ID)
	if err != nil {
		t.Fatalf("CancelReservation failed: %v", err)
	}

	// Verify via GetMemberReservations - cancelled reservation should still be in list
	reservations, _ := reservationSvc.GetMemberReservations(member2.ID)
	var found *models.Reservation
	for _, r := range reservations {
		if r.ID == res.ID && r.Status == models.ReservationStatusCancelled {
			found = r
			break
		}
	}
	if found == nil {
		t.Error("expected to find cancelled reservation")
	}
}

// getReservationRepo is a test helper - in production we'd use dependency injection
// For this test we verify via GetMemberReservations
func TestGetMemberReservations(t *testing.T) {
	reservationSvc, loanSvc, bookRepo, memberRepo := setupReservationTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)

	book, _ := libSvc.AddBook("Book", "Author", "666", "Subject", 1)
	member1, _ := libSvc.RegisterMember("Alice", "alice@test.com", models.MembershipStandard)
	member2, _ := libSvc.RegisterMember("Bob", "bob@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member1.ID)
	res, _ := reservationSvc.Reserve(book.ID, member2.ID)

	reservations, err := reservationSvc.GetMemberReservations(member2.ID)
	if err != nil {
		t.Fatalf("GetMemberReservations failed: %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("expected 1 reservation, got %d", len(reservations))
	}
	if reservations[0].ID != res.ID {
		t.Error("wrong reservation returned")
	}
}
