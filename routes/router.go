package routes

import (
	"InfinityInitiative/routes/characters"
	"InfinityInitiative/routes/game"
	"InfinityInitiative/routes/items"
	"InfinityInitiative/routes/levels"
	"InfinityInitiative/routes/monsters"
	"InfinityInitiative/routes/recipes"
	"InfinityInitiative/routes/roles"
	"InfinityInitiative/routes/skills"
	"InfinityInitiative/routes/users"

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
	roles.SetupEndpoints(openGroup)
	levels.SetupEndpoints(openGroup)
	skills.SetupEndpoints(openGroup)
	items.SetupEndpoints(openGroup)
	recipes.SetupEndpoints(openGroup)
	characters.SetupEndpoints(openGroup)
	monsters.SetupEndpoints(openGroup)
}
