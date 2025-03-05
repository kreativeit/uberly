package ride

import (
	"errors"
	"time"
)

type Ride struct {
	RideId    string    `json:"rideId"`
	RiderId   string    `json:"riderId"`
	DriverId  string    `json:"driverId"`
	Status    string    `json:"status"`
	Location  Location  `json:"location"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Events chan<- any `json:"-"`
}

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
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
