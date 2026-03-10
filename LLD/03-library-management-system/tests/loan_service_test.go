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

func setupLoanTest(t *testing.T) (*services.LoanService, interfaces.BookRepository, interfaces.MemberRepository) {
	bookRepo := repositories.NewInMemoryBookRepo()
	memberRepo := repositories.NewInMemoryMemberRepo()
	loanRepo := repositories.NewInMemoryLoanRepo()
	reservationRepo := repositories.NewInMemoryReservationRepo()
	fineRepo := repositories.NewInMemoryFineRepo()
	notifMgr := notifications.NewNotificationManager()
	notifMgr.Register(notifications.NewConsoleNotifier())
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

	return loanSvc, bookRepo, memberRepo
}

func TestCheckOut_Success(t *testing.T) {
	loanSvc, bookRepo, memberRepo := setupLoanTest(t)

	// Add book and member via library service
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	book, _ := libSvc.AddBook("Test Book", "Author", "123", "Subject", 2)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	loan, err := loanSvc.CheckOut(book.ID, member.ID)
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}
	if loan.Status != models.LoanStatusActive {
		t.Errorf("expected status Active, got %s", loan.Status)
	}
	if loan.BookID != book.ID || loan.MemberID != member.ID {
		t.Error("loan has wrong book or member ID")
	}

	// Verify book available copies decreased
	updatedBook, _ := bookRepo.GetByID(book.ID)
	if updatedBook.AvailableCopies != 1 {
		t.Errorf("expected 1 available copy, got %d", updatedBook.AvailableCopies)
	}
}

func TestCheckOut_NoAvailableCopies(t *testing.T) {
	loanSvc, bookRepo, memberRepo := setupLoanTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	book, _ := libSvc.AddBook("Single Copy", "Author", "456", "Subject", 1)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member.ID)
	_, err := loanSvc.CheckOut(book.ID, member.ID)
	if err != services.ErrBookNotAvailable {
		t.Errorf("expected ErrBookNotAvailable, got %v", err)
	}
}

func TestCheckOut_MemberLimitReached(t *testing.T) {
	loanSvc, bookRepo, memberRepo := setupLoanTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	book1, _ := libSvc.AddBook("Book1", "A", "1", "S", 1)
	book2, _ := libSvc.AddBook("Book2", "A", "2", "S", 1)
	book3, _ := libSvc.AddBook("Book3", "A", "3", "S", 1)
	book4, _ := libSvc.AddBook("Book4", "A", "4", "S", 1)
	book5, _ := libSvc.AddBook("Book5", "A", "5", "S", 1)
	book6, _ := libSvc.AddBook("Book6", "A", "6", "S", 1)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	// Borrow 5 books (max limit)
	for _, b := range []*models.Book{book1, book2, book3, book4, book5} {
		_, _ = loanSvc.CheckOut(b.ID, member.ID)
	}
	_, err := loanSvc.CheckOut(book6.ID, member.ID)
	if err != services.ErrMemberCannotBorrow {
		t.Errorf("expected ErrMemberCannotBorrow, got %v", err)
	}
}

func TestReturn_Success(t *testing.T) {
	loanSvc, bookRepo, memberRepo := setupLoanTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	book, _ := libSvc.AddBook("Return Test", "Author", "789", "Subject", 1)
	member, _ := libSvc.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member.ID)
	returned, err := loanSvc.Return(book.ID, member.ID)
	if err != nil {
		t.Fatalf("Return failed: %v", err)
	}
	if returned.Status != models.LoanStatusReturned {
		t.Errorf("expected status Returned, got %s", returned.Status)
	}
	if returned.ReturnDate == nil {
		t.Error("ReturnDate should be set")
	}

	updatedBook, _ := bookRepo.GetByID(book.ID)
	if updatedBook.AvailableCopies != 1 {
		t.Errorf("expected 1 available copy after return, got %d", updatedBook.AvailableCopies)
	}
}

func TestReturn_WrongMember(t *testing.T) {
	loanSvc, bookRepo, memberRepo := setupLoanTest(t)
	libSvc := services.NewLibraryService(bookRepo, memberRepo)
	book, _ := libSvc.AddBook("Book", "A", "999", "S", 1)
	member1, _ := libSvc.RegisterMember("Alice", "a@test.com", models.MembershipStandard)
	member2, _ := libSvc.RegisterMember("Bob", "b@test.com", models.MembershipStandard)

	_, _ = loanSvc.CheckOut(book.ID, member1.ID)
	_, err := loanSvc.Return(book.ID, member2.ID)
	if err != services.ErrBookNotCheckedOut {
		t.Errorf("expected ErrBookNotCheckedOut, got %v", err)
	}
}

func TestReturn_OverdueCreatesFine(t *testing.T) {
	// We need to add fineRepo to the loan service - but setupLoanTest uses its own fineRepo
	// The loan service's Return creates fine via fineRepo - so we need access to it
	// Let's create a custom setup that returns fineRepo
	bookRepo2 := repositories.NewInMemoryBookRepo()
	memberRepo2 := repositories.NewInMemoryMemberRepo()
	loanRepo := repositories.NewInMemoryLoanRepo()
	reservationRepo := repositories.NewInMemoryReservationRepo()
	fineRepo2 := repositories.NewInMemoryFineRepo()
	notifMgr := notifications.NewNotificationManager()
	fineCalc := services.NewPerDayFineCalculator()

	loanSvc2 := services.NewLoanService(services.LoanServiceConfig{
		BookRepo:        bookRepo2,
		MemberRepo:      memberRepo2,
		LoanRepo:        loanRepo,
		ReservationRepo: reservationRepo,
		FineRepo:        fineRepo2,
		FineCalculator:  fineCalc,
		Notifier:        notifMgr,
	})

	libSvc2 := services.NewLibraryService(bookRepo2, memberRepo2)
	book, _ := libSvc2.AddBook("Overdue Return", "Author", "777", "Subject", 1)
	member, _ := libSvc2.RegisterMember("Test", "test@test.com", models.MembershipStandard)

	loan, _ := loanSvc2.CheckOut(book.ID, member.ID)
	// Manually backdate the loan to simulate overdue (DueDate 10 days ago)
	loan.DueDate = time.Now().Add(-10 * 24 * time.Hour)
	_ = loanRepo.Update(loan)

	_, err := loanSvc2.Return(book.ID, member.ID)
	if err != nil {
		t.Fatalf("Return failed: %v", err)
	}

	fines, _ := fineRepo2.GetPendingByMemberID(member.ID)
	if len(fines) != 1 {
		t.Fatalf("expected 1 fine after overdue return, got %d", len(fines))
	}
	if fines[0].Amount != 10.0 {
		t.Errorf("expected fine $10 (10 days), got $%.2f", fines[0].Amount)
	}
}
