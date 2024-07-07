package roles

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/roles", GetAll)
	g.GET("/roles/:id", EditRole)
	g.POST("/roles", SaveRole)
	g.GET("/roles/create", CreateRole)
	g.GET("/roles/delete/:id", ConfirmDeleteRole)
	g.DELETE("/roles/:id", DeleteRole)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/roles/*.html",
	})

	roles := models.GetAllRoles()

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       roles,
		ActivePage: "roles",
		Character:  models.GetCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditRole(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/roles/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	role := models.GetRole(Id)

	data := models.Data{
		Page: role,
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveRole(c echo.Context) error {
	role := new(models.Role)

	if err := c.Bind(role); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if role.ID != 0 {
		_, err = role.Update()
	} else {
		_, err = role.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Role")
	}

	c.Response().Header().Set("HX-Redirect", "/roles")
	c.Response().WriteHeader(300)

	return nil
}

func CreateRole(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/roles/*.html",
	})

	return c.Render(http.StatusOK, "table-edit", nil)
}

func ConfirmDeleteRole(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/roles/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	role := models.GetRole(Id)

	data := models.Data{
		Page: role,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteRole(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	role := models.GetRole(Id)

	_, err := role.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Role")
	}

	c.Response().Header().Set("HX-Redirect", "/roles")
	c.Response().WriteHeader(300)

	return nil
}
