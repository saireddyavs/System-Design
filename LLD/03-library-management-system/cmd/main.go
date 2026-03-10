// Package main demonstrates the Library Management System
package main

import (
	"fmt"
	"library-management-system/internal/models"
	"library-management-system/internal/notifications"
	"library-management-system/internal/repositories"
	"library-management-system/internal/services"
	"time"
)

func main() {
	// Initialize repositories (DIP - inject concrete implementations)
	bookRepo := repositories.NewInMemoryBookRepo()
	memberRepo := repositories.NewInMemoryMemberRepo()
	loanRepo := repositories.NewInMemoryLoanRepo()
	reservationRepo := repositories.NewInMemoryReservationRepo()
	fineRepo := repositories.NewInMemoryFineRepo()

	// Notification manager with observers
	notifMgr := notifications.NewNotificationManager()
	notifMgr.Register(notifications.NewConsoleNotifier())
	notifMgr.Register(notifications.NewEmailNotifier())

	// Fine calculator strategy ($1/day)
	fineCalc := services.NewPerDayFineCalculator()

	// Initialize services
	librarySvc := services.NewLibraryService(bookRepo, memberRepo)
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
	fineSvc := services.NewFineService(services.FineServiceConfig{
		FineRepo:       fineRepo,
		LoanRepo:       loanRepo,
		FineCalculator: fineCalc,
		Notifier:       notifMgr,
		MemberRepo:     memberRepo,
		BookRepo:       bookRepo,
	})
	searchSvc := services.NewSearchService(bookRepo)
	reminderSvc := services.NewDueDateReminderService(loanRepo, memberRepo, bookRepo, notifMgr)

	// Demo flow
	fmt.Println("=== Library Management System Demo ===")
	fmt.Println()

	// Add books
	book1, _ := librarySvc.AddBook("Clean Code", "Robert Martin", "978-0132350884", "Programming", 2)
	book2, _ := librarySvc.AddBook("Design Patterns", "Gang of Four", "978-0201633610", "Software Engineering", 1)
	fmt.Printf("Added books: %s, %s\n\n", book1.Title, book2.Title)

	// Register members
	member1, _ := librarySvc.RegisterMember("Alice", "alice@lib.com", models.MembershipStandard)
	member2, _ := librarySvc.RegisterMember("Bob", "bob@lib.com", models.MembershipStandard)
	fmt.Printf("Registered members: %s, %s\n\n", member1.Name, member2.Name)

	// Search
	results, _ := searchSvc.Search(services.SearchCriteria{Author: "Martin"})
	fmt.Printf("Search by author 'Martin': found %d book(s)\n", len(results))
	for _, b := range results {
		fmt.Printf("  - %s by %s\n", b.Title, b.Author)
	}
	fmt.Println()

	// Check out
	loan1, _ := loanSvc.CheckOut(book1.ID, member1.ID)
	fmt.Printf("Alice checked out: %s (due: %s)\n\n", book1.Title, loan1.DueDate.Format("2006-01-02"))

	// Bob reserves (book has 1 copy left, or if both checked out)
	_, _ = loanSvc.CheckOut(book2.ID, member1.ID)
	_, _ = reservationSvc.Reserve(book2.ID, member2.ID)
	fmt.Printf("Bob reserved: %s (Alice has it)\n\n", book2.Title)

	// Return book2 - Bob gets notified
	_, _ = loanSvc.Return(book2.ID, member1.ID)
	fmt.Println("Alice returned Design Patterns -> Bob notified")
	fmt.Println()

	// Process overdue (simulate - in real scenario would have overdue loans)
	_, _ = fineSvc.ProcessOverdueLoans()

	// Due date reminders
	count, _ := reminderSvc.SendRemindersForLoansDueWithin(14 * 24 * time.Hour)
	fmt.Printf("Sent %d due date reminder(s)\n\n", count)

	fmt.Println("=== Demo Complete ===")
}
