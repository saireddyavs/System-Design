package services

import (
	"hotel-management-system/internal/interfaces"
	"hotel-management-system/internal/models"
)

// PaymentService handles payment operations
type PaymentService struct {
	paymentRepo interfaces.PaymentRepository
	processor   interfaces.PaymentProcessor
}

// NewPaymentService creates a new payment service
func NewPaymentService(paymentRepo interfaces.PaymentRepository, processor interfaces.PaymentProcessor) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		processor:   processor,
	}
}

// ProcessPayment processes a payment for a booking
func (s *PaymentService) ProcessPayment(payment *models.Payment) error {
	if payment.GetStatus() == models.PaymentStatusCompleted {
		return ErrPaymentAlreadyPaid
	}
	if err := s.processor.ProcessPayment(payment); err != nil {
		return err
	}
	return s.paymentRepo.Update(payment)
}

// ProcessRefund processes a refund
func (s *PaymentService) ProcessRefund(payment *models.Payment, amount float64) error {
	if err := s.processor.ProcessRefund(payment, amount); err != nil {
		return err
	}
	return s.paymentRepo.Update(payment)
}
