package manager

import (
	"context"
	"log"
	"time"
	"uberly/ride"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RideManager struct {
	redis      *redis.Client
	rideEvents chan any
}

type LocationUpdate struct {
	ride.Location
	Id   string
	Type string // Either rider or driver
}

func NewRideManager() *RideManager {
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB,
	})

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := redis.Ping(timeout).Err(); err != nil {
		log.Fatal(err)
	}

	rideEvents := make(chan any, 100)

	go func() {
		for event := range rideEvents {
			// TODO: Persist to database
			log.Println(event)
		}
	}()

	return &RideManager{
		redis:      redis,
		rideEvents: rideEvents,
	}
}

func (m *RideManager) NewRide(at ride.Location, riderId string) *ride.Ride {
	r := &ride.Ride{
		Location:  at,
		RiderId:   riderId,
		Status:    "created",
		CreatedAt: time.Now(),
		Events:    m.rideEvents,
		RideId:    "ride:" + uuid.New().String(),
	}

	m.rideEvents <- ride.NewRideCreatedEvent(r)

	return r
}

func (s *RideManager) Start(ride *ride.Ride) {
	stage := NewPendingStage(s.redis)

	if err := stage.handle(ride); err != nil {
		log.Println(err)
		return
	}
}

func (s *RideManager) UpdateLocation(location LocationUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	key := "drivers"
	if location.Type == "rider" {
		key = "riders"
	}

	s.redis.GeoAdd(ctx, key, &redis.GeoLocation{
		Name:      location.Id,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
	})
}
