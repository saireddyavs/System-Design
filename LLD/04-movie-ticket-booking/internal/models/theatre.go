package models

// Theatre represents a theatre with multiple screens
type Theatre struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	City    string   `json:"city"`
	Address string   `json:"address"`
	Screens []Screen `json:"screens,omitempty"`
}
