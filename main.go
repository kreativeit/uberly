package main

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"time"
	"uberly/ride"
	"uberly/ride/manager"
	"uberly/server"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	app := echo.New()
	m := manager.NewRideManager()

	go driverWorker(m, "driver:mock", time.Second*30)
	go driverWorker(m, "driver-mock", time.Second*10)

	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	app.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := server.ServerContext{
				Context:     c,
				RideManager: m,
			}

			return next(ctx)
		}
	})

	app.GET("/", func(c echo.Context) error {
		at := ride.Location{
			Latitude:  0,
			Longitude: 0,
		}

		manager := c.(server.ServerContext).RideManager
		ride := manager.NewRide(at, "rider:mock")

		go manager.Start(ride)

		return c.JSON(http.StatusOK, ride)
	})

	app.Start(":3000")
}

func driverWorker(s *manager.RideManager, driverId string, delay time.Duration) {
	r := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB,
	})

	dChannel := r.Subscribe(context.Background(), driverId).Channel()

	go func() {
		for {
			s.UpdateLocation(manager.LocationUpdate{
				Id:   driverId,
				Type: "driver",
				Location: ride.Location{
					Latitude:  rand.Float64(),
					Longitude: rand.Float64(),
				},
			})

			time.Sleep(time.Second * 5)
		}
	}()

	go func() {
		for message := range dChannel {
			var payload *manager.RideRequestCmd

			if err := json.Unmarshal([]byte(message.Payload), &payload); err == nil {
				reply := manager.RideAcceptCmd{
					DriverId:   driverId,
					RideId:     payload.RideId,
					AcceptedAt: time.Now(),
				}

				// Simulating a slow reply
				time.Sleep(delay)

				r.Publish(context.Background(), payload.RideId, reply)
			}
		}
	}()
}
