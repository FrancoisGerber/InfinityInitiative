package main

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"InfinityInitiative/routes"

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

	// Static files and templates
	e.Static("/assets", "views/assets")

	// Router
	routes.Router(e)

	// Connect to database
	data.Init()
	migrations()

	// Start Server
	e.Logger.Fatal(e.Start(":8069"))
}

// Database auto migrations
func migrations() {
	models.MigrateUser()
}
