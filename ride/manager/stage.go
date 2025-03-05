package manager

import "uberly/ride"

type RideStage interface {
	handle(ride *ride.Ride) error
}
