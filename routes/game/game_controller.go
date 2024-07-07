package game

import (
	"InfinityInitiative/data/models"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/game/:skillName", GamePage)

}

func GamePage(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	skillName := c.Param("skillName")
	data := models.Data{
		Page:       skillName,
		ActivePage: skillName,
	}

	return c.Render(http.StatusOK, "index.html", data)
}
