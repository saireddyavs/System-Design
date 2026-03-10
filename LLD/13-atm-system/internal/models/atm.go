package models

import (
	"sync"
	"time"
)

// CashInventory represents the cash available in ATM by denomination
type CashInventory map[Denomination]int

// ATM represents the ATM machine
type ATM struct {
	mu            sync.RWMutex
	ID            string
	Location      string
	CashAvailable CashInventory
	TotalCash     float64
	State         ATMState
	CurrentCardID string
	CurrentPIN    string
	LastActivity  time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewATM creates a new ATM instance
func NewATM(id, location string, cashAvailable CashInventory) *ATM {
	totalCash := 0.0
	for denom, count := range cashAvailable {
		totalCash += float64(denom) * float64(count)
	}
	now := time.Now()
	return &ATM{
		ID:            id,
		Location:      location,
		CashAvailable: cashAvailable,
		TotalCash:     totalCash,
		State:         ATMStateIdle,
		LastActivity:  now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GetState returns current ATM state (thread-safe)
func (a *ATM) GetState() ATMState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}

// SetState updates ATM state (thread-safe)
func (a *ATM) SetState(state ATMState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = state
	a.UpdatedAt = time.Now()
	a.LastActivity = time.Now()
}

// GetTotalCash returns total cash available (thread-safe)
func (a *ATM) GetTotalCash() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.TotalCash
}

// UpdateCashInventory updates cash after dispensing (thread-safe)
func (a *ATM) UpdateCashInventory(dispensed map[Denomination]int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for denom, count := range dispensed {
		a.CashAvailable[denom] -= count
		a.TotalCash -= float64(denom) * float64(count)
	}
	a.UpdatedAt = time.Now()
}

// GetCashInventory returns a copy of cash inventory (thread-safe)
func (a *ATM) GetCashInventory() CashInventory {
	a.mu.RLock()
	defer a.mu.RUnlock()
	inventory := make(CashInventory)
	for k, v := range a.CashAvailable {
		inventory[k] = v
	}
	return inventory
}

// AddCash adds cash to inventory (e.g., after deposit)
func (a *ATM) AddCash(denom Denomination, count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.CashAvailable[denom] += count
	a.TotalCash += float64(denom) * float64(count)
	a.UpdatedAt = time.Now()
}

// IsOutOfService checks if ATM has insufficient cash
func (a *ATM) IsOutOfService() bool {
	return a.GetTotalCash() < CashThreshold
}
