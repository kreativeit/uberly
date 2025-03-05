package ride

import "time"

type RideEvent struct {
	Type      string
	RideId    string
	CreatedAt time.Time
}

type RideAcceptedEvent struct {
	RideEvent
	DriverId string
}

func NewRideCreatedEvent(ride *Ride) RideEvent {
	return RideEvent{
		Type:      "Created",
		RideId:    ride.RideId,
		CreatedAt: time.Now(),
	}
}

func NewRideAcceptedEvent(ride *Ride) RideAcceptedEvent {
	return RideAcceptedEvent{
		DriverId: ride.DriverId,
		RideEvent: RideEvent{
			Type:      "Accepted",
			RideId:    ride.RideId,
			CreatedAt: time.Now(),
		},
	}
}

func NewDriversNotAvailable(ride *Ride) RideEvent {
	return RideEvent{
		Type:      "Timeout",
		CreatedAt: time.Now(),
		RideId:    ride.RideId,
	}
}
