package auth

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"InfinityInitiative/views"
	"github.com/labstack/echo/v4"
	"net/http"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/login", Login)
	g.POST("/login", PostLogin)
	g.GET("/logout", Logout)
	g.GET("/register", Register)
	g.POST("/register", PostRegister)
}

func Login(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/auth/*.html",
		"views/auth/login/*.html",
	})

	return c.Render(http.StatusOK, "login.html", nil)
}

func PostLogin(c echo.Context) error {
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	dbUser := models.GetUserByEmail(user.Email)

	if dbUser.ID == 0 || dbUser.Password == "" {
		return c.String(http.StatusBadRequest, "Email not registered!")
	}

	verified := utils.VerifyPassword(user.Password, dbUser.Password)

	if !verified {
		return c.String(http.StatusUnauthorized, "Password is incorrect!")
	}

	err := utils.CreateSessionUser(c, dbUser.ID)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}

func Logout(c echo.Context) error {
	utils.CreateSessionUser(c, 0)

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func Register(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/auth/*.html",
		"views/auth/register/*.html",
	})

	return c.Render(http.StatusOK, "register.html", nil)
}

func PostRegister(c echo.Context) error {
	user := new(models.User)

	if err := c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	user.Password = utils.HashPassword(user.Password)

	_, err := user.Create()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save User")
	}

	c.Response().Header().Set("HX-Redirect", "/login")
	c.Response().WriteHeader(300)

	return nil
}
