package models

// Airport represents an airport with IATA code and location details
type Airport struct {
	Code    string // IATA 3-letter code (e.g., JFK, LAX)
	Name    string
	City    string
	Country string
}

// NewAirport creates a new Airport instance
func NewAirport(code, name, city, country string) *Airport {
	return &Airport{
		Code:    code,
		Name:    name,
		City:    city,
		Country: country,
	}
}
