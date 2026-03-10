package services

import "errors"

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrRoomNotAvailable     = errors.New("room not available for dates")
	ErrGuestNotFound        = errors.New("guest not found")
	ErrBookingNotFound      = errors.New("booking not found")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrInvalidDateRange     = errors.New("check-out must be after check-in")
	ErrPaymentRequired     = errors.New("payment required before confirmation")
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrPaymentAlreadyPaid  = errors.New("payment already completed")
)
