package models

import "sync"

// LoyaltyTier represents guest loyalty level for pricing discounts
type LoyaltyTier string

const (
	LoyaltyTierNone    LoyaltyTier = "None"
	LoyaltyTierSilver  LoyaltyTier = "Silver"
	LoyaltyTierGold    LoyaltyTier = "Gold"
	LoyaltyTierPlatinum LoyaltyTier = "Platinum"
)

// LoyaltyTierThresholds defines points needed for each tier
var LoyaltyTierThresholds = map[LoyaltyTier]int{
	LoyaltyTierNone:    0,
	LoyaltyTierSilver:  1000,
	LoyaltyTierGold:    5000,
	LoyaltyTierPlatinum: 15000,
}

// LoyaltyTierDiscounts defines discount percentage per tier
var LoyaltyTierDiscounts = map[LoyaltyTier]float64{
	LoyaltyTierNone:    0,
	LoyaltyTierSilver:  0.05,  // 5%
	LoyaltyTierGold:    0.10,  // 10%
	LoyaltyTierPlatinum: 0.15, // 15%
}

// Guest represents a hotel guest
type Guest struct {
	ID            string
	Name          string
	Email         string
	Phone         string
	IDProof       string
	LoyaltyPoints int
	mu            sync.RWMutex
}

// NewGuest creates a new Guest instance
func NewGuest(id, name, email, phone, idProof string) *Guest {
	return &Guest{
		ID:            id,
		Name:          name,
		Email:         email,
		Phone:         phone,
		IDProof:       idProof,
		LoyaltyPoints: 0,
	}
}

// GetLoyaltyPoints returns loyalty points (thread-safe)
func (g *Guest) GetLoyaltyPoints() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LoyaltyPoints
}

// AddLoyaltyPoints adds points (thread-safe)
func (g *Guest) AddLoyaltyPoints(points int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.LoyaltyPoints += points
}

// GetLoyaltyTier returns the tier based on points
func (g *Guest) GetLoyaltyTier() LoyaltyTier {
	points := g.GetLoyaltyPoints()
	if points >= LoyaltyTierThresholds[LoyaltyTierPlatinum] {
		return LoyaltyTierPlatinum
	}
	if points >= LoyaltyTierThresholds[LoyaltyTierGold] {
		return LoyaltyTierGold
	}
	if points >= LoyaltyTierThresholds[LoyaltyTierSilver] {
		return LoyaltyTierSilver
	}
	return LoyaltyTierNone
}
