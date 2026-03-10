package models

import (
	"sync"
	"time"
)

// Account represents a bank account
type Account struct {
	mu                sync.RWMutex
	ID                string
	AccountNumber     string
	HolderName        string
	Type              AccountType
	Balance           float64
	PINHash           string // In production, use bcrypt/argon2
	DailyWithdrawn    float64
	LastWithdrawalDate time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewAccount creates a new account
func NewAccount(id, accountNumber, holderName string, accountType AccountType, balance float64, pinHash string) *Account {
	now := time.Now()
	return &Account{
		ID:                id,
		AccountNumber:     accountNumber,
		HolderName:        holderName,
		Type:              accountType,
		Balance:           balance,
		PINHash:           pinHash,
		DailyWithdrawn:    0,
		LastWithdrawalDate: time.Time{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// GetBalance returns the current balance (thread-safe)
func (a *Account) GetBalance() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Balance
}

// AddBalance adds amount to balance (thread-safe)
func (a *Account) AddBalance(amount float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Balance += amount
	a.UpdatedAt = time.Now()
}

// DeductBalance deducts amount from balance (thread-safe)
func (a *Account) DeductBalance(amount float64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Balance < amount {
		return false
	}
	a.Balance -= amount
	a.UpdatedAt = time.Now()
	return true
}

// GetDailyWithdrawn returns daily withdrawn amount (thread-safe)
func (a *Account) GetDailyWithdrawn() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.DailyWithdrawn
}

// RecordWithdrawal records a withdrawal for daily limit tracking
func (a *Account) RecordWithdrawal(amount float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	if a.LastWithdrawalDate.YearDay() != now.YearDay() || a.LastWithdrawalDate.Year() != now.Year() {
		a.DailyWithdrawn = 0
		a.LastWithdrawalDate = now
	}
	a.DailyWithdrawn += amount
	a.UpdatedAt = now
}

// VerifyPIN checks if the provided PIN matches (simplified - production would use bcrypt)
func (a *Account) VerifyPIN(pin string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.PINHash == pin // In production: bcrypt.CompareHashAndPassword
}

// UpdatePIN updates the account PIN
func (a *Account) UpdatePIN(newPINHash string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.PINHash = newPINHash
	a.UpdatedAt = time.Now()
}
