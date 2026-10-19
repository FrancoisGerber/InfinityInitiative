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

// RoleAdmin is the role id allowed to reach the content-management screens.
const RoleAdmin uint = 1

func Router(e *echo.Echo) {
	// Public: login, register, logout.
	openGroup := e.Group("")
	auth.SetupEndpoints(openGroup)

	// Authenticated: play the game.
	sessionGroup := e.Group("")
	sessionGroup.Use(utils.SessionAuth)
	game.SetupEndpoints(sessionGroup)

	// Admin: content management. Requires a logged-in user *with the admin role* —
	// hiding links in the sidebar is not authorisation.
	adminGroup := sessionGroup.Group("")
	adminGroup.Use(requireRole(RoleAdmin))
	users.SetupEndpoints(adminGroup)
	roles.SetupEndpoints(adminGroup)
	levels.SetupEndpoints(adminGroup)
	skills.SetupEndpoints(adminGroup)
	items.SetupEndpoints(adminGroup)
	recipes.SetupEndpoints(adminGroup)
	characters.SetupEndpoints(adminGroup)
	monsters.SetupEndpoints(adminGroup)
}
