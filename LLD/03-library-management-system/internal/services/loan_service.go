package services

import (
	"errors"
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"time"

	"github.com/google/uuid"
)

const LoanPeriodDays = 14

var (
	ErrBookNotAvailable   = errors.New("book has no available copies")
	ErrMemberCannotBorrow = errors.New("member cannot borrow (inactive or limit reached)")
	ErrLoanNotFound      = errors.New("loan not found")
	ErrBookNotCheckedOut = errors.New("book is not checked out by this member")
)

// LoanService handles lending and returning (SRP - Factory for loan creation)
type LoanService struct {
	bookRepo        interfaces.BookRepository
	memberRepo      interfaces.MemberRepository
	loanRepo        interfaces.LoanRepository
	reservationRepo interfaces.ReservationRepository
	fineRepo        interfaces.FineRepository
	fineCalculator  *PerDayFineCalculator
	notifier        interfaces.NotifyBroadcaster
}

// LoanServiceConfig holds dependencies for LoanService (DIP - depends on interfaces)
type LoanServiceConfig struct {
	BookRepo        interfaces.BookRepository
	MemberRepo      interfaces.MemberRepository
	LoanRepo        interfaces.LoanRepository
	ReservationRepo interfaces.ReservationRepository
	FineRepo        interfaces.FineRepository
	FineCalculator  *PerDayFineCalculator
	Notifier        interfaces.NotifyBroadcaster
}

// NewLoanService creates a new loan service
func NewLoanService(cfg LoanServiceConfig) *LoanService {
	return &LoanService{
		bookRepo:        cfg.BookRepo,
		memberRepo:      cfg.MemberRepo,
		loanRepo:        cfg.LoanRepo,
		reservationRepo: cfg.ReservationRepo,
		fineRepo:        cfg.FineRepo,
		fineCalculator:  cfg.FineCalculator,
		notifier:        cfg.Notifier,
	}
}

// CheckOut creates a new loan (Factory: creates Loan with proper defaults)
func (s *LoanService) CheckOut(bookID, memberID string) (*models.Loan, error) {
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil {
		return nil, err
	}
	if !book.HasAvailableCopies() {
		return nil, ErrBookNotAvailable
	}

	member, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, err
	}
	if !member.CanBorrow() {
		return nil, ErrMemberCannotBorrow
	}

	now := time.Now()
	dueDate := now.AddDate(0, 0, LoanPeriodDays)

	loan := &models.Loan{
		ID:        uuid.New().String(),
		BookID:    bookID,
		MemberID:  memberID,
		IssueDate: now,
		DueDate:   dueDate,
		Status:    models.LoanStatusActive,
		CreatedAt: now,
	}
	if err := s.loanRepo.Create(loan); err != nil {
		return nil, err
	}

	book.AvailableCopies--
	if book.AvailableCopies == 0 {
		book.Status = models.BookStatusCheckedOut
	}
	_ = s.bookRepo.Update(book)

	member.BorrowedCount++
	_ = s.memberRepo.Update(member)

	return loan, nil
}

// Return processes book return, creates fine if overdue, notifies first reservation
func (s *LoanService) Return(bookID, memberID string) (*models.Loan, error) {
	loan, err := s.loanRepo.GetActiveByBookID(bookID)
	if err != nil || loan.MemberID != memberID {
		return nil, ErrBookNotCheckedOut
	}

	book, _ := s.bookRepo.GetByID(bookID)
	member, _ := s.memberRepo.GetByID(memberID)

	now := time.Now()
	loan.ReturnDate = &now
	loan.Status = models.LoanStatusReturned
	_ = s.loanRepo.Update(loan)

	book.AvailableCopies++
	book.Status = models.BookStatusAvailable
	_ = s.bookRepo.Update(book)

	member.BorrowedCount--
	_ = s.memberRepo.Update(member)

	// Create fine if overdue
	if now.After(loan.DueDate) {
		amount := s.fineCalculator.Calculate(loan, now)
		if amount > 0 {
			fine := &models.Fine{
				ID:        uuid.New().String(),
				LoanID:    loan.ID,
				MemberID:  memberID,
				Amount:    amount,
				Status:    models.FineStatusPending,
				CreatedAt: now,
			}
			_ = s.fineRepo.Create(fine)
		}
	}

	// Notify first person in reservation queue
	s.notifyFirstReservation(book)

	return loan, nil
}

func (s *LoanService) notifyFirstReservation(book *models.Book) {
	reservations, _ := s.reservationRepo.GetPendingByBookID(book.ID)
	if len(reservations) == 0 {
		return
	}
	first := reservations[0]
	resMember, _ := s.memberRepo.GetByID(first.MemberID)
	now := time.Now()
	first.Status = models.ReservationStatusNotified
	first.NotifiedAt = &now
	first.ExpiresAt = now.AddDate(0, 0, 3)
	_ = s.reservationRepo.Update(first)

	s.notifier.NotifyAll(interfaces.NotificationPayload{
		Type:        interfaces.NotificationReservationReady,
		MemberID:    resMember.ID,
		MemberEmail: resMember.Email,
		BookTitle:   book.Title,
		BookID:      book.ID,
		Message:     "Your reserved book '" + book.Title + "' is now available. Please pick it up within 3 days.",
	})
}

// SendRemindersForLoansDueWithin sends reminders for loans due within the given duration
func (s *LoanService) SendRemindersForLoansDueWithin(within time.Duration) (int, error) {
	deadline := time.Now().Add(within)
	loans, err := s.loanRepo.GetLoansDueBefore(deadline)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	count := 0
	for _, loan := range loans {
		if loan.Status != models.LoanStatusActive {
			continue
		}
		if now.After(loan.DueDate) {
			continue // Skip overdue - handled by FineService
		}
		member, _ := s.memberRepo.GetByID(loan.MemberID)
		book, _ := s.bookRepo.GetByID(loan.BookID)
		if member == nil || book == nil {
			continue
		}
		s.notifier.NotifyAll(interfaces.NotificationPayload{
			Type:        interfaces.NotificationDueDateReminder,
			MemberID:    member.ID,
			MemberEmail: member.Email,
			BookTitle:   book.Title,
			BookID:      book.ID,
			DueDate:     loan.DueDate,
			Message:     "Reminder: Your book '" + book.Title + "' is due on " + loan.DueDate.Format("2006-01-02") + ". Please return it on time.",
		})
		count++
	}
	return count, nil
}
