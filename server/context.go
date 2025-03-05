package server

import (
	"uberly/ride/manager"

	"github.com/labstack/echo/v4"
)

type ServerContext struct {
	echo.Context
	RideManager *manager.RideManager
}
