package models

import "sync"

var (
	instance *ParkingLot
	once     sync.Once
)

// ParkingLot is the singleton managing all parking levels and spots.
// Singleton pattern: ensures single global instance for the entire parking system.
type ParkingLot struct {
	mu      sync.RWMutex
	ID      string
	Name    string
	Levels  []*ParkingLevel
	levelID map[string]*ParkingLevel
}

// GetInstance returns the singleton ParkingLot instance.
// Thread-safe initialization via sync.Once.
func GetInstance() *ParkingLot {
	once.Do(func() {
		instance = &ParkingLot{
			ID:      "PL-001",
			Name:    "Main Parking Lot",
			Levels:  []*ParkingLevel{},
			levelID: make(map[string]*ParkingLevel),
		}
	})
	return instance
}

// Initialize configures the parking lot with levels and spots.
// Must be called before use. Safe to call multiple times; re-initializes.
func (p *ParkingLot) Initialize(levels []*ParkingLevel) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Levels = levels
	p.levelID = make(map[string]*ParkingLevel)
	for _, l := range levels {
		p.levelID[l.ID] = l
	}
}

// GetLevel returns a level by ID.
func (p *ParkingLot) GetLevel(levelID string) *ParkingLevel {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.levelID[levelID]
}

// GetLevels returns all levels.
func (p *ParkingLot) GetLevels() []*ParkingLevel {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]*ParkingLevel{}, p.Levels...)
}

// ResetInstance resets the singleton for testing. Not for production use.
func ResetInstance() {
	instance = nil
	once = sync.Once{}
}
