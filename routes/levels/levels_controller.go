package levels

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/levels", GetAll)
	g.GET("/levels/:id", EditLevel)
	g.POST("/levels", SaveLevel)
	g.GET("/levels/create", CreateLevel)
	g.GET("/levels/delete/:id", ConfirmDeleteLevel)
	g.DELETE("/levels/:id", DeleteLevel)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/levels/*.html",
	})

	levels := models.GetAllLevels()

	data := models.Data{
		Page:       levels,
		ActivePage: "levels",
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditLevel(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/levels/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	level := models.GetLevel(Id)

	data := models.Data{
		Page: level,
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveLevel(c echo.Context) error {
	level := new(models.Level)

	if err := c.Bind(level); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if level.ID != 0 {
		_, err = level.Update()
	} else {
		_, err = level.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Level")
	}

	c.Response().Header().Set("HX-Redirect", "/levels")
	c.Response().WriteHeader(300)

	return nil
}

func CreateLevel(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/levels/*.html",
	})

	return c.Render(http.StatusOK, "table-edit", nil)
}

func ConfirmDeleteLevel(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/levels/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	level := models.GetLevel(Id)

	data := models.Data{
		Page: level,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteLevel(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	level := models.GetLevel(Id)

	_, err := level.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Level")
	}

	c.Response().Header().Set("HX-Redirect", "/levels")
	c.Response().WriteHeader(300)

	return nil
}
