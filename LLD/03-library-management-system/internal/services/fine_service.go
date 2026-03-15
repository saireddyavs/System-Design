package services

import (
	"errors"
	"fmt"
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFineNotFound   = errors.New("fine not found")
	ErrFineAlreadyPaid = errors.New("fine already paid")
)

// FineService handles overdue fines (SRP)
type FineService struct {
	fineRepo       interfaces.FineRepository
	loanRepo       interfaces.LoanRepository
	fineCalculator *PerDayFineCalculator
	notifier       interfaces.NotifyBroadcaster
	memberRepo     interfaces.MemberRepository
	bookRepo       interfaces.BookRepository
}

// FineServiceConfig holds dependencies
type FineServiceConfig struct {
	FineRepo       interfaces.FineRepository
	LoanRepo       interfaces.LoanRepository
	FineCalculator *PerDayFineCalculator
	Notifier       interfaces.NotifyBroadcaster
	MemberRepo     interfaces.MemberRepository
	BookRepo       interfaces.BookRepository
}

// NewFineService creates a new fine service
func NewFineService(cfg FineServiceConfig) *FineService {
	return &FineService{
		fineRepo:       cfg.FineRepo,
		loanRepo:       cfg.LoanRepo,
		fineCalculator: cfg.FineCalculator,
		notifier:       cfg.Notifier,
		memberRepo:     cfg.MemberRepo,
		bookRepo:       cfg.BookRepo,
	}
}

// ProcessOverdueLoans creates fines for overdue loans and sends notifications
func (s *FineService) ProcessOverdueLoans() ([]*models.Fine, error) {
	overdueLoans, err := s.loanRepo.GetOverdueLoans()
	if err != nil {
		return nil, err
	}

	var createdFines []*models.Fine
	now := time.Now()

	for _, loan := range overdueLoans {
		existingFine, _ := s.fineRepo.GetByLoanID(loan.ID)
		if existingFine != nil {
			continue
		}

		amount := s.fineCalculator.Calculate(loan, now)
		if amount <= 0 {
			continue
		}

		fine := &models.Fine{
			ID:        uuid.New().String(),
			LoanID:    loan.ID,
			MemberID:  loan.MemberID,
			Amount:    amount,
			Status:    models.FineStatusPending,
			CreatedAt: now,
		}
		if err := s.fineRepo.Create(fine); err != nil {
			continue
		}
		createdFines = append(createdFines, fine)

		// Send overdue notification
		member, _ := s.memberRepo.GetByID(loan.MemberID)
		book, _ := s.bookRepo.GetByID(loan.BookID)
		if member != nil && book != nil {
			s.notifier.NotifyAll(interfaces.NotificationPayload{
				Type:        interfaces.NotificationOverdue,
				MemberID:    member.ID,
				MemberEmail: member.Email,
				BookTitle:   book.Title,
				BookID:      book.ID,
				DueDate:     loan.DueDate,
				Message:     "Your book '" + book.Title + "' is overdue. Fine: $" + fmt.Sprintf("%.2f", amount),
			})
		}
	}
	return createdFines, nil
}

// PayFine marks a fine as paid
func (s *FineService) PayFine(fineID string) error {
	fine, err := s.fineRepo.GetByID(fineID)
	if err != nil {
		return err
	}
	if fine.Status == models.FineStatusPaid {
		return ErrFineAlreadyPaid
	}
	now := time.Now()
	fine.Status = models.FineStatusPaid
	fine.PaidAt = &now
	return s.fineRepo.Update(fine)
}

// GetMemberPendingFines returns all pending fines for a member
func (s *FineService) GetMemberPendingFines(memberID string) ([]*models.Fine, error) {
	return s.fineRepo.GetPendingByMemberID(memberID)
}
