package models

import "math"

// Location represents a geographic coordinate
type Location struct {
	Latitude  float64
	Longitude float64
}

// HaversineDistance calculates the distance in kilometers between two locations
// using the Haversine formula for great-circle distance
func HaversineDistance(from, to Location) float64 {
	const earthRadius = 6371 // km

	lat1 := from.Latitude * math.Pi / 180
	lat2 := to.Latitude * math.Pi / 180
	deltaLat := (to.Latitude - from.Latitude) * math.Pi / 180
	deltaLon := (to.Longitude - from.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
