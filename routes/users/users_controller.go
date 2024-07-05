package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/", GetAll)

}

func GetAll(c echo.Context) error {
	return c.Render(http.StatusOK, "index", nil)
}
