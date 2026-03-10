package models

import "time"

// Passenger represents a passenger/customer in the system
type Passenger struct {
	ID             string
	Name           string
	Email          string
	Phone          string
	PassportNumber string
	DateOfBirth    time.Time
}

// NewPassenger creates a new Passenger instance
func NewPassenger(id, name, email, phone, passportNumber string, dateOfBirth time.Time) *Passenger {
	return &Passenger{
		ID:             id,
		Name:           name,
		Email:          email,
		Phone:          phone,
		PassportNumber: passportNumber,
		DateOfBirth:    dateOfBirth,
	}
}
