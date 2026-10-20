package characters

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/characters/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/characters", GetAll)
	g.GET("/characters/:id", EditCharacter)
	g.POST("/characters", SaveCharacter)
	g.GET("/characters/create", CreateCharacter)
	g.GET("/characters/delete/:id", ConfirmDeleteCharacter)
	g.DELETE("/characters/:id", DeleteCharacter)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	characters := character.GetAllCharacters()

	data := models.Data{
		Page:       characters,
		ActivePage: "characters",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditCharacter(c echo.Context) error {

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

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveCharacter(c echo.Context) error {
	characterDto := new(models.Character)
	if err := c.Bind(characterDto); err != nil {
		utils.ShowAlertError(c, "Could not read the form. Reload the page and try again.")
		return nil
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
		utils.ShowAlertError(c, "Could Not Save Character")
		return nil
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}

func CreateCharacter(c echo.Context) error {

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

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func ConfirmDeleteCharacter(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	character := models.NewCharacter()
	character = character.GetCharacter(Id)

	data := models.Data{
		Page: character,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteCharacter(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	character := models.NewCharacter()
	character = character.GetCharacter(Id)

	_, err := character.Delete()

	if err != nil {
		utils.ShowAlertError(c, "Could Not Delete Character")
		return nil
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}
