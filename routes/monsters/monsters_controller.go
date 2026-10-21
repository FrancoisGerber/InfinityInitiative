package monsters

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/monsters/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/monsters", GetAll)
	g.GET("/monsters/:id", EditMonster)
	g.POST("/monsters", SaveMonster)
	g.GET("/monsters/create", CreateMonster)
	g.GET("/monsters/delete/:id", ConfirmDeleteMonster)
	g.DELETE("/monsters/:id", DeleteMonster)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	monster := models.NewMonster()
	monsters := monster.GetAllMonsters()
	character := models.NewCharacter()

	data := models.Data{
		Page:       monsters,
		ActivePage: "monsters",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditMonster(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.NewMonster()
	level := models.NewLevel()
	monster = monster.GetMonster(Id)
	levels := level.GetAllLevels()

	data := models.Data{
		Page: monster,
		DropdownData: map[string]any{
			"Levels":  levels,
			"Classes": models.Classes,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveMonster(c echo.Context) error {
	monsterDto := new(models.Monster)

	if err := c.Bind(monsterDto); err != nil {
		return utils.ShowAlertError(c, http.StatusBadRequest, "Could not read the form. Reload the page and try again.")
	}

	var err error
	monster := models.InitMonster(monsterDto)

	if monster.ID != 0 {
		_, err = monster.Update()
	} else {
		_, err = monster.Create()
	}

	if err != nil {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could Not Save Monster")
	}

	c.Response().Header().Set("HX-Redirect", "/monsters")
	c.Response().WriteHeader(300)

	return nil
}

func CreateMonster(c echo.Context) error {

	level := models.NewLevel()
	levels := level.GetAllLevels()

	data := models.Data{
		Page: models.Monster{},
		DropdownData: map[string]any{
			"Levels":  levels,
			"Classes": models.Classes,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func ConfirmDeleteMonster(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.NewMonster()
	monster = monster.GetMonster(Id)

	data := models.Data{
		Page: monster,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteMonster(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	monster := models.NewMonster()
	monster = monster.GetMonster(Id)

	_, err := monster.Delete()

	if err != nil {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could Not Delete Monster")
	}

	c.Response().Header().Set("HX-Redirect", "/monsters")
	c.Response().WriteHeader(300)

	return nil
}
