package characters

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

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
	c.Echo().Renderer = templates.For("views/layout/*.html","views/characters/*.html")

	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	characters := character.GetAllCharacters()

	data := models.Data{
		Page:       characters,
		ActivePage: "characters",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditCharacter(c echo.Context) error {
	c.Echo().Renderer = templates.For("views/layout/*.html","views/characters/*.html")

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	character := models.NewCharacter()
	level := models.NewLevel()
	user := models.NewUser()

	character = character.GetCharacter(Id)
	levels := level.GetAllLevels()
	users := user.GetAllUsers()

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
	characterDto := new(models.Character)
	if err := c.Bind(characterDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	character := models.InitCharacter(characterDto)

	character.Image = "/assets/img/classes/" + character.Class + ".png"

	var err error

	if character.ID != 0 {
		_, err = character.Update()
	} else {
		_, err = character.Create()
	}

	user := models.NewUser()
	user = user.GetUser(int64(character.UserId))

	if user.ID != 0 {
		user.Image = character.Image
		_, err = user.Update()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Character")
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}

func CreateCharacter(c echo.Context) error {
	c.Echo().Renderer = templates.For("views/layout/*.html","views/characters/*.html")

	level := models.NewLevel()
	user := models.NewUser()

	levels := level.GetAllLevels()
	users := user.GetAllUsers()

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
	c.Echo().Renderer = templates.For("views/layout/*.html","views/characters/*.html")

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	character := models.NewCharacter()
	character = character.GetCharacter(Id)

	data := models.Data{
		Page: character,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteCharacter(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	character := models.NewCharacter()
	character = character.GetCharacter(Id)

	_, err := character.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Character")
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}
