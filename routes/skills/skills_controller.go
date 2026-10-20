package skills

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/skills/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/skills", GetAll)
	g.GET("/skills/:id", EditSkill)
	g.POST("/skills", SaveSkill)
	g.GET("/skills/create", CreateSkill)
	g.GET("/skills/delete/:id", ConfirmDeleteSkill)
	g.DELETE("/skills/:id", DeleteSkill)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	skill := models.NewSkill()
	skills := skill.GetAllSkills()
	character := models.NewCharacter()

	data := models.Data{
		Page:       skills,
		ActivePage: "skills",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditSkill(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.NewSkill()
	skill = skill.GetSkill(Id)

	data := models.Data{
		Page: skill,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveSkill(c echo.Context) error {
	skillDto := new(models.Skill)

	if err := c.Bind(skillDto); err != nil {
		utils.ShowAlertError(c, "Could not read the form. Reload the page and try again.")
		return nil
	}

	var err error
	skill := models.InitSkill(skillDto)

	if skill.ID != 0 {
		_, err = skill.Update()
	} else {
		_, err = skill.Create()
	}

	if err != nil {
		utils.ShowAlertError(c, "Could Not Save Skill")
		return nil
	}

	c.Response().Header().Set("HX-Redirect", "/skills")
	c.Response().WriteHeader(300)

	return nil
}

func CreateSkill(c echo.Context) error {

	return templates.RenderGroup(c, http.StatusOK, "table-edit", nil, viewPatterns...)
}

func ConfirmDeleteSkill(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.NewSkill()
	skill = skill.GetSkill(Id)

	data := models.Data{
		Page: skill,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteSkill(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.NewSkill()
	skill = skill.GetSkill(Id)

	_, err := skill.Delete()

	if err != nil {
		utils.ShowAlertError(c, "Could Not Delete Skill")
		return nil
	}

	c.Response().Header().Set("HX-Redirect", "/skills")
	c.Response().WriteHeader(300)

	return nil
}
