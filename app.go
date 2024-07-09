package main

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"InfinityInitiative/routes"
	"InfinityInitiative/utils"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Setup Echo
	e := echo.New()

	// Root level middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(session.Middleware(sessions.NewCookieStore(utils.Secret)))

	// Static files and templates
	e.Static("/assets", "views/assets")

	// Router
	routes.Router(e)

	// Run DB migrations
	//TODO: Remove db init as the models do this themselves
	data.Init(utils.DBName)
	autoMigrations()

	// Start Server
	e.Logger.Fatal(e.Start(":8069"))
}

// Database auto migrations
func autoMigrations() {
	character := models.NewCharacter(utils.DBName)
	data.AutoMigration(&character)

	//TODO: Refactor all models to implement Model interface and change migrations
	models.MigrateRoles()
	models.MigrateUser()
	models.MigrateLevels()
	models.MigrateItems()
	models.MigrateSkills()
	models.MigrateRecipes()
	models.MigrateMonsters()
	models.MigrateInventorys()
	models.MigrateCharacterSkills()
}
