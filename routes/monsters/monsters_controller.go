package monsters

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/monsters", GetAll)
	g.GET("/monsters/:id", EditMonster)
	g.POST("/monsters", SaveMonster)
	g.GET("/monsters/create", CreateMonster)
	g.GET("/monsters/delete/:id", ConfirmDeleteMonster)
	g.DELETE("/monsters/:id", DeleteMonster)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/monsters/*.html",
	})

	userId := c.Get("UserId").(uint)

	monsters := models.GetAllMonsters()
	character := models.NewCharacter()

	data := models.Data{
		Page:       monsters,
		ActivePage: "monsters",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditMonster(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/monsters/*.html",
	})

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.GetMonster(Id)
	levels := models.GetAllLevels()

	data := models.Data{
		Page: monster,
		DropdownData: map[string]any{
			"Levels":  levels,
			"Classes": models.Classes,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveMonster(c echo.Context) error {
	monster := new(models.Monster)

	if err := c.Bind(monster); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if monster.ID != 0 {
		_, err = monster.Update()
	} else {
		_, err = monster.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Monster")
	}

	c.Response().Header().Set("HX-Redirect", "/monsters")
	c.Response().WriteHeader(300)

	return nil
}

func CreateMonster(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/monsters/*.html",
	})

	levels := models.GetAllLevels()

	data := models.Data{
		Page: models.Monster{},
		DropdownData: map[string]any{
			"Levels":  levels,
			"Classes": models.Classes,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func ConfirmDeleteMonster(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/monsters/*.html",
	})

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.GetMonster(Id)

	data := models.Data{
		Page: monster,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteMonster(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.GetMonster(Id)

	_, err := monster.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Monster")
	}

	c.Response().Header().Set("HX-Redirect", "/monsters")
	c.Response().WriteHeader(300)

	return nil
}
