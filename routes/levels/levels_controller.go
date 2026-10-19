package levels

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

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
	c.Echo().Renderer = templates.For("views/layout/*.html","views/levels/*.html")

	userId := c.Get("UserId").(uint)

	level := models.NewLevel()
	levels := level.GetAllLevels()
	character := models.NewCharacter()

	data := models.Data{
		Page:       levels,
		ActivePage: "levels",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditLevel(c echo.Context) error {
	c.Echo().Renderer = templates.For("views/layout/*.html","views/levels/*.html")

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	level := models.NewLevel()
	level = level.GetLevel(Id)

	data := models.Data{
		Page: level,
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveLevel(c echo.Context) error {
	levelDto := new(models.Level)

	if err := c.Bind(levelDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error
	level := models.InitLevel(levelDto)

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
	c.Echo().Renderer = templates.For("views/layout/*.html","views/levels/*.html")

	return c.Render(http.StatusOK, "table-edit", nil)
}

func ConfirmDeleteLevel(c echo.Context) error {
	c.Echo().Renderer = templates.For("views/layout/*.html","views/levels/*.html")

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	level := models.NewLevel()
	level = level.GetLevel(Id)

	data := models.Data{
		Page: level,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteLevel(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	level := models.NewLevel()
	level = level.GetLevel(Id)

	_, err := level.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Level")
	}

	c.Response().Header().Set("HX-Redirect", "/levels")
	c.Response().WriteHeader(300)

	return nil
}
