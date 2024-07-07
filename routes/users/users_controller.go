package users

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/users", GetAll)
	g.GET("/users/:id", EditUser)
	g.POST("/users", SaveUser)
	g.GET("/users/create", CreateUser)
	g.GET("/users/delete/:id", ConfirmDeleteUser)
	g.DELETE("/users/:id", DeleteUser)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/users/*.html",
	})

	users := models.GetAllUsers()

	data := models.Data{
		Page:       users,
		ActivePage: "users",
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditUser(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/users/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	user := models.GetUser(Id)

	data := models.Data{
		Page: user,
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveUser(c echo.Context) error {
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

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
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/users/*.html",
	})

	return c.Render(http.StatusOK, "table-edit", nil)
}

func ConfirmDeleteUser(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/users/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	user := models.GetUser(Id)

	data := models.Data{
		Page: user,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteUser(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	user := models.GetUser(Id)

	_, err := user.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete User")
	}

	c.Response().Header().Set("HX-Redirect", "/users")
	c.Response().WriteHeader(300)

	return nil
}
