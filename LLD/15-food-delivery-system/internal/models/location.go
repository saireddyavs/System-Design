package models

import "math"

// Location represents geographic coordinates
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// EarthRadiusKm is the Earth's radius in kilometers for Haversine formula
const EarthRadiusKm = 6371.0

// Distance calculates the Haversine distance between two locations in kilometers
func (l Location) Distance(other Location) float64 {
	lat1 := l.Lat * math.Pi / 180
	lat2 := other.Lat * math.Pi / 180
	deltaLat := (other.Lat - l.Lat) * math.Pi / 180
	deltaLng := (other.Lng - l.Lng) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}
