package roles

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/roles/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/roles", GetAll)
	g.GET("/roles/:id", EditRole)
	g.POST("/roles", SaveRole)
	g.GET("/roles/create", CreateRole)
	g.GET("/roles/delete/:id", ConfirmDeleteRole)
	g.DELETE("/roles/:id", DeleteRole)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	role := models.NewRole()
	roles := role.GetAllRoles()
	character := models.NewCharacter()

	data := models.Data{
		Page:       roles,
		ActivePage: "roles",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditRole(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	role := models.NewRole()
	role = role.GetRole(Id)

	data := models.Data{
		Page: role,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveRole(c echo.Context) error {
	roleDto := new(models.Role)

	if err := c.Bind(roleDto); err != nil {
		return utils.ShowAlertError(c, http.StatusBadRequest, "Could not read the form. Reload the page and try again.")
	}

	var err error
	role := models.InitRole(roleDto)

	if role.ID != 0 {
		_, err = role.Update()
	} else {
		_, err = role.Create()
	}

	if err != nil {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could Not Save Role")
	}

	c.Response().Header().Set("HX-Redirect", "/roles")
	c.Response().WriteHeader(300)

	return nil
}

func CreateRole(c echo.Context) error {

	return templates.RenderGroup(c, http.StatusOK, "table-edit", nil, viewPatterns...)
}

func ConfirmDeleteRole(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	role := models.NewRole()
	role = role.GetRole(Id)

	data := models.Data{
		Page: role,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteRole(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	role := models.NewRole()
	role = role.GetRole(Id)

	_, err := role.Delete()

	if err != nil {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could Not Delete Role")
	}

	c.Response().Header().Set("HX-Redirect", "/roles")
	c.Response().WriteHeader(300)

	return nil
}
