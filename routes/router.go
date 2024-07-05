package routes

import (
	"InfinityInitiative/routes/users"

	"InfinityInitiative/routes/game"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func Router(e *echo.Echo) {
	// Router groups
	authGroup := e.Group("/admin")
	openGroup := e.Group("")

	// Auth
	authGroup.Use(echojwt.JWT([]byte("InfinityInitiative")))

	//Routes
	game.SetupEndpoints(openGroup)
	users.SetupEndpoints(openGroup)
}
