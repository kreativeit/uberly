package manager

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	"uberly/ride"

	"github.com/redis/go-redis/v9"
)

type PendingStage struct {
	SearchOptions
	next  *RideStage
	redis *redis.Client
}

type SearchOptions struct {
	radius   float64
	attempts int8
	ctx      context.Context
}

type RideRequestCmd struct {
	RideId   string
	Location ride.Location
}

func (i RideRequestCmd) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type RideAcceptCmd struct {
	RideId     string
	DriverId   string
	AcceptedAt time.Time
}

func (i RideAcceptCmd) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func NewPendingStage(redis *redis.Client) *PendingStage {
	ctx := context.WithValue(context.Background(), "attempts", 0)

	return &PendingStage{
		redis: redis,
		next:  nil,
		SearchOptions: SearchOptions{
			ctx:      ctx,
			radius:   5,
			attempts: 0,
		},
	}
}

func (stage *PendingStage) handle(r *ride.Ride) error {
	ctx := stage.ctx

	query := &redis.GeoRadiusQuery{
		WithDist:  true,
		WithCoord: true,
		Sort:      "ASC",
		Radius:    stage.radius,
	}

	attempts := ctx.Value("attempts").(int)

	if attempts >= 10 {
		return errors.New("exceeded max attempts")
	}

	drivers, err := stage.redis.GeoRadius(ctx, "drivers", r.Location.Longitude, r.Location.Latitude, query).Result()

	if err != nil {
		return err
	}

	if len(drivers) == 0 {
		stage.radius *= 2
		stage.attempts += 1
		return stage.handle(r)
	}

	rChannel := stage.redis.Subscribe(context.Background(), r.RideId).Channel()

	for _, driver := range drivers {
		// Give 25 seconds for driver to reply
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*25)
		defer cancel()

		request := RideRequestCmd{
			RideId:   r.RideId,
			Location: r.Location,
		}

		if err := stage.redis.Publish(context.Background(), driver.Name, request).Err(); err != nil {
			continue
		}

		select {
		case message := <-rChannel:
			var payload *RideAcceptCmd

			if err := json.Unmarshal([]byte(message.Payload), &payload); err == nil {
				if payload.DriverId == driver.Name {
					r.AssignDriver(driver.Name)
					log.Println(r)
				}

				return nil
			}

		case <-ctx.Done():
			// TODO: Handle expired request.
			// If request expires and there are no drivers left, notify user.
			r.Events <- ride.NewDriversNotAvailable(r)

			continue
		}

	}

	return nil
}
