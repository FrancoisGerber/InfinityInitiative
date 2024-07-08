package skills

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/skills", GetAll)
	g.GET("/skills/:id", EditSkill)
	g.POST("/skills", SaveSkill)
	g.GET("/skills/create", CreateSkill)
	g.GET("/skills/delete/:id", ConfirmDeleteSkill)
	g.DELETE("/skills/:id", DeleteSkill)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/skills/*.html",
	})

	skills := models.GetAllSkills()

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       skills,
		ActivePage: "skills",
		Character:  models.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditSkill(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/skills/*.html",
	})

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.GetSkill(Id)

	data := models.Data{
		Page: skill,
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveSkill(c echo.Context) error {
	skill := new(models.Skill)

	if err := c.Bind(skill); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if skill.ID != 0 {
		_, err = skill.Update()
	} else {
		_, err = skill.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Skill")
	}

	c.Response().Header().Set("HX-Redirect", "/skills")
	c.Response().WriteHeader(300)

	return nil
}

func CreateSkill(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/skills/*.html",
	})

	return c.Render(http.StatusOK, "table-edit", nil)
}

func ConfirmDeleteSkill(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/skills/*.html",
	})

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.GetSkill(Id)

	data := models.Data{
		Page: skill,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteSkill(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	skill := models.GetSkill(Id)

	_, err := skill.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Skill")
	}

	c.Response().Header().Set("HX-Redirect", "/skills")
	c.Response().WriteHeader(300)

	return nil
}
