package characters

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/characters", GetAll)
	g.GET("/characters/:id", EditCharacter)
	g.POST("/characters", SaveCharacter)
	g.GET("/characters/create", CreateCharacter)
	g.GET("/characters/delete/:id", ConfirmDeleteCharacter)
	g.DELETE("/characters/:id", DeleteCharacter)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/characters/*.html",
	})

	characters := models.GetAllCharacters()

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       characters,
		ActivePage: "characters",
		Character:  models.GetCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditCharacter(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/characters/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	character := models.GetCharacter(Id)
	levels := models.GetAllLevels()
	users := models.GetAllUsers()

	data := models.Data{
		Page: character,
		DropdownData: map[string]any{
			"Levels":  levels,
			"Users":   users,
			"Classes": models.Classes,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveCharacter(c echo.Context) error {
	character := new(models.Character)

	if err := c.Bind(character); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if character.ID != 0 {
		_, err = character.Update()
	} else {
		_, err = character.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Character")
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}

func CreateCharacter(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/characters/*.html",
	})

	levels := models.GetAllLevels()
	users := models.GetAllUsers()

	data := models.Data{
		Page: models.Character{},
		DropdownData: map[string]any{
			"Levels":  levels,
			"Users":   users,
			"Classes": models.Classes,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func ConfirmDeleteCharacter(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/characters/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	character := models.GetCharacter(Id)

	data := models.Data{
		Page: character,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteCharacter(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	character := models.GetCharacter(Id)

	_, err := character.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Character")
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}
