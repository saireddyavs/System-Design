package models

// SeatClass represents the class/cabin of a seat
type SeatClass string

const (
	SeatClassEconomy  SeatClass = "Economy"
	SeatClassBusiness SeatClass = "Business"
	SeatClassFirst    SeatClass = "First"
)

// SeatStatus represents the availability status of a seat
type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "Available"
	SeatStatusBooked    SeatStatus = "Booked"
	SeatStatusBlocked   SeatStatus = "Blocked"
)

// FlightStatus represents the operational status of a flight
type FlightStatus string

const (
	FlightStatusScheduled FlightStatus = "Scheduled"
	FlightStatusOnTime    FlightStatus = "OnTime"
	FlightStatusDelayed   FlightStatus = "Delayed"
	FlightStatusCancelled FlightStatus = "Cancelled"
	FlightStatusDeparted  FlightStatus = "Departed"
	FlightStatusArrived   FlightStatus = "Arrived"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "Confirmed"
	BookingStatusCancelled  BookingStatus = "Cancelled"
	BookingStatusPending   BookingStatus = "Pending"
	BookingStatusCheckedIn BookingStatus = "CheckedIn"
)

// BaggageAllowance returns the baggage allowance in kg for each seat class
func (s SeatClass) BaggageAllowance() int {
	switch s {
	case SeatClassEconomy:
		return 23
	case SeatClassBusiness:
		return 32
	case SeatClassFirst:
		return 40
	default:
		return 23
	}
}

// ClassMultiplier returns the price multiplier for each seat class
func (s SeatClass) ClassMultiplier() float64 {
	switch s {
	case SeatClassEconomy:
		return 1.0
	case SeatClassBusiness:
		return 2.5
	case SeatClassFirst:
		return 5.0
	default:
		return 1.0
	}
}
