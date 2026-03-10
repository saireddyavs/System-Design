package models

// SpotSize represents the size category of a parking spot.
// Used for compatibility matching: a spot can hold vehicles of its size or smaller.
type SpotSize int

const (
	SpotSizeSmall SpotSize = iota
	SpotSizeMedium
	SpotSizeLarge
)

func (s SpotSize) String() string {
	return [...]string{"Small", "Medium", "Large"}[s]
}

// VehicleType represents the type of vehicle for parking compatibility.
type VehicleType int

const (
	VehicleTypeMotorcycle VehicleType = iota
	VehicleTypeCar
	VehicleTypeBus
)

func (v VehicleType) String() string {
	return [...]string{"Motorcycle", "Car", "Bus"}[v]
}

// Vehicle defines the interface all vehicle types must implement.
// LSP: All concrete vehicles (Motorcycle, Car, Bus) can be substituted
// wherever Vehicle is expected without breaking behavior.
type Vehicle interface {
	GetType() VehicleType
	GetLicensePlate() string
	// Returns the minimum spot size this vehicle requires
	GetRequiredSpotSize() SpotSize
}

// BaseVehicle provides common vehicle fields. Used by concrete implementations.
type BaseVehicle struct {
	LicensePlate string
	Type         VehicleType
}

// GetLicensePlate returns the vehicle's license plate.
func (v *BaseVehicle) GetLicensePlate() string {
	return v.LicensePlate
}

// GetType returns the vehicle type.
func (v *BaseVehicle) GetType() VehicleType {
	return v.Type
}

// Motorcycle requires Small spot size.
type Motorcycle struct {
	BaseVehicle
}

// GetRequiredSpotSize returns SpotSizeSmall - motorcycles fit in smallest spots.
func (m *Motorcycle) GetRequiredSpotSize() SpotSize {
	return SpotSizeSmall
}

// Car requires Medium spot size.
type Car struct {
	BaseVehicle
}

// GetRequiredSpotSize returns SpotSizeMedium - cars need medium or larger spots.
func (c *Car) GetRequiredSpotSize() SpotSize {
	return SpotSizeMedium
}

// Bus requires Large spot size.
type Bus struct {
	BaseVehicle
}

// GetRequiredSpotSize returns SpotSizeLarge - buses need large spots only.
func (b *Bus) GetRequiredSpotSize() SpotSize {
	return SpotSizeLarge
}

// VehicleFactory creates vehicles by type. Factory pattern for encapsulation.
// OCP: new vehicle types can be added by extending this factory without
// modifying existing vehicle creation logic.
func NewVehicle(vehicleType VehicleType, licensePlate string) Vehicle {
	switch vehicleType {
	case VehicleTypeMotorcycle:
		return &Motorcycle{BaseVehicle{LicensePlate: licensePlate, Type: vehicleType}}
	case VehicleTypeCar:
		return &Car{BaseVehicle{LicensePlate: licensePlate, Type: vehicleType}}
	case VehicleTypeBus:
		return &Bus{BaseVehicle{LicensePlate: licensePlate, Type: vehicleType}}
	default:
		return &Car{BaseVehicle{LicensePlate: licensePlate, Type: VehicleTypeCar}}
	}
}
