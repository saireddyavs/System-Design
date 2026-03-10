package services

import (
	"errors"
	"library-management-system/internal/interfaces"
	"library-management-system/internal/models"
	"time"

	"github.com/google/uuid"
)

var (
	ErrBookAvailable       = errors.New("book has available copies - no need to reserve")
	ErrMemberAlreadyReserved = errors.New("member already has a pending reservation for this book")
	ErrReservationNotFound  = errors.New("reservation not found")
)

// ReservationService handles book reservations (SRP)
type ReservationService struct {
	bookRepo        interfaces.BookRepository
	memberRepo      interfaces.MemberRepository
	loanRepo        interfaces.LoanRepository
	reservationRepo interfaces.ReservationRepository
}

// NewReservationService creates a new reservation service
func NewReservationService(
	bookRepo interfaces.BookRepository,
	memberRepo interfaces.MemberRepository,
	loanRepo interfaces.LoanRepository,
	reservationRepo interfaces.ReservationRepository,
) *ReservationService {
	return &ReservationService{
		bookRepo:        bookRepo,
		memberRepo:      memberRepo,
		loanRepo:        loanRepo,
		reservationRepo: reservationRepo,
	}
}

// Reserve creates a reservation for a checked-out book
func (s *ReservationService) Reserve(bookID, memberID string) (*models.Reservation, error) {
	book, err := s.bookRepo.GetByID(bookID)
	if err != nil {
		return nil, err
	}
	if book.HasAvailableCopies() {
		return nil, ErrBookAvailable
	}

	_, err = s.memberRepo.GetByID(memberID)
	if err != nil {
		return nil, err
	}

	// Check if member already has pending reservation for this book
	existing, _ := s.reservationRepo.GetByMemberID(memberID)
	for _, r := range existing {
		if r.BookID == bookID && r.Status == models.ReservationStatusPending {
			return nil, ErrMemberAlreadyReserved
		}
	}

	now := time.Now()
	reservation := &models.Reservation{
		ID:         uuid.New().String(),
		BookID:     bookID,
		MemberID:   memberID,
		ReservedAt: now,
		Status:     models.ReservationStatusPending,
		ExpiresAt:  now.AddDate(0, 0, 30), // 30 days to fulfill
		CreatedAt:  now,
	}
	if err := s.reservationRepo.Create(reservation); err != nil {
		return nil, err
	}
	return reservation, nil
}

// CancelReservation cancels a pending reservation
func (s *ReservationService) CancelReservation(reservationID, memberID string) error {
	res, err := s.reservationRepo.GetByID(reservationID)
	if err != nil {
		return err
	}
	if res.MemberID != memberID {
		return errors.New("reservation does not belong to member")
	}
	if res.Status != models.ReservationStatusPending {
		return errors.New("can only cancel pending reservations")
	}
	res.Status = models.ReservationStatusCancelled
	return s.reservationRepo.Update(res)
}

// GetMemberReservations returns all reservations for a member
func (s *ReservationService) GetMemberReservations(memberID string) ([]*models.Reservation, error) {
	return s.reservationRepo.GetByMemberID(memberID)
}
