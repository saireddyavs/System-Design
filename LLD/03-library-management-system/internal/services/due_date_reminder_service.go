package services

import (
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"time"
)

// DueDateReminderService sends due date reminders (Observer trigger)
// Typically called by a scheduler/cron job
type DueDateReminderService struct {
	loanRepo   interfaces.LoanRepository
	memberRepo interfaces.MemberRepository
	bookRepo   interfaces.BookRepository
	notifier   interfaces.NotifyBroadcaster
}

// NewDueDateReminderService creates a new reminder service
func NewDueDateReminderService(
	loanRepo interfaces.LoanRepository,
	memberRepo interfaces.MemberRepository,
	bookRepo interfaces.BookRepository,
	notifier interfaces.NotifyBroadcaster,
) *DueDateReminderService {
	return &DueDateReminderService{
		loanRepo:   loanRepo,
		memberRepo: memberRepo,
		bookRepo:   bookRepo,
		notifier:   notifier,
	}
}

// SendRemindersForLoansDueWithin sends reminders for loans due within the given duration
func (s *DueDateReminderService) SendRemindersForLoansDueWithin(within time.Duration) (int, error) {
	deadline := time.Now().Add(within)
	loans, err := s.loanRepo.GetLoansDueBefore(deadline)
	if err != nil {
		return 0, err
	}

	// Filter to only active loans due in the future (not yet overdue)
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
