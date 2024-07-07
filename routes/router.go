package routes

import (
	"InfinityInitiative/routes/auth"
	"InfinityInitiative/routes/characters"
	"InfinityInitiative/routes/game"
	"InfinityInitiative/routes/items"
	"InfinityInitiative/routes/levels"
	"InfinityInitiative/routes/monsters"
	"InfinityInitiative/routes/recipes"
	"InfinityInitiative/routes/roles"
	"InfinityInitiative/routes/skills"
	"InfinityInitiative/routes/users"
	"InfinityInitiative/utils"

	"github.com/labstack/echo/v4"
)

func Router(e *echo.Echo) {
	// Router groups
	closedGroup := e.Group("")
	openGroup := e.Group("")

	// Auth
	closedGroup.Use(utils.SessionAuth)

	//Routes
	auth.SetupEndpoints(openGroup)
	game.SetupEndpoints(closedGroup)
	users.SetupEndpoints(closedGroup)
	roles.SetupEndpoints(closedGroup)
	levels.SetupEndpoints(closedGroup)
	skills.SetupEndpoints(closedGroup)
	items.SetupEndpoints(closedGroup)
	recipes.SetupEndpoints(closedGroup)
	characters.SetupEndpoints(closedGroup)
	monsters.SetupEndpoints(closedGroup)
}
