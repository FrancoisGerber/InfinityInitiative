package main

import (
	"InfinityInitiative/routes"
	"InfinityInitiative/views"

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
	e.Static("/assets", "views/static")
	e.Renderer = templates.ParseTemplates()

	// Router
	routes.Router(e)

	// Start Server
	e.Logger.Fatal(e.Start(":8069"))
}
