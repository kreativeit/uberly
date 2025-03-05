package ride

import (
	"errors"
	"time"
)

type Ride struct {
	RideId    string
	RiderId   string
	DriverId  string
	Status    string
	Location  Location
	CreatedAt time.Time
	UpdatedAt time.Time

	Events chan<- any
}

type Location struct {
	Latitude, Longitude float64
}

func (r *Ride) AssignDriver(driverId string) error {
	if r.DriverId != "" {
		return errors.New("ride already accepted")
	}

	r.DriverId = driverId
	r.Status = "accepted"
	r.Events <- NewRideAcceptedEvent(r)

	return nil
}
