package users

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/users/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/users", GetAll)
	g.GET("/users/:id", EditUser)
	g.POST("/users", SaveUser)
	g.GET("/users/create", CreateUser)
	g.GET("/users/delete/:id", ConfirmDeleteUser)
	g.DELETE("/users/:id", DeleteUser)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	user := models.NewUser()
	character := models.NewCharacter()

	users := user.GetAllUsers()

	data := models.Data{
		Page:       users,
		ActivePage: "users",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditUser(c echo.Context) error {

	id := c.Param("id")
	Id, _ := strconv.Atoi(id)

	user := models.NewUser()
	user = user.GetUser(int64(Id))
	user.Password = ""

	data := models.Data{
		Page: user,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveUser(c echo.Context) error {
	userDto := new(models.User)

	if err := c.Bind(userDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error
	user := models.InitUser(userDto)

	user.Password = utils.HashPassword(user.Password)

	if user.ID != 0 {
		_, err = user.Update()
	} else {
		_, err = user.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save User")
	}

	c.Response().Header().Set("HX-Redirect", "/users")
	c.Response().WriteHeader(300)

	return nil
}

func CreateUser(c echo.Context) error {

	return templates.RenderGroup(c, http.StatusOK, "table-edit", nil, viewPatterns...)
}

func ConfirmDeleteUser(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	user := models.NewUser()
	user = user.GetUser(Id)

	data := models.Data{
		Page: user,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteUser(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	user := models.NewUser()
	user = user.GetUser(Id)

	_, err := user.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete User")
	}

	c.Response().Header().Set("HX-Redirect", "/users")
	c.Response().WriteHeader(300)

	return nil
}
