package auth

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"InfinityInitiative/views"
	"github.com/labstack/echo/v4"
	"net/http"
)

// viewPatterns for the auth screens: the shells plus the per-screen fragments.
var (
	authViewPatterns     = "views/auth/*.html"
	loginViewPatterns    = "views/auth/login/*.html"
	registerViewPatterns = "views/auth/register/*.html"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/login", Login)
	g.POST("/login", PostLogin)
	g.GET("/logout", Logout)
	g.GET("/register", Register)
	g.POST("/register", PostRegister)
	g.GET("/clearalert", ClearAlert)
}

func Login(c echo.Context) error {
	return templates.RenderGroup(c, http.StatusOK, "login.html", nil, authViewPatterns, loginViewPatterns)
}

func PostLogin(c echo.Context) error {
	userDto := new(models.User)

	if err := c.Bind(userDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	user := models.InitUser(userDto)

	dbUser := user.GetUserByEmail(user.Email)

	if dbUser.ID == 0 || dbUser.Password == "" {
		return c.String(http.StatusBadRequest, "Email not registered!")
	}

	verified := utils.VerifyPassword(user.Password, dbUser.Password)

	if !verified {
		return c.String(http.StatusUnauthorized, "Password is incorrect!")
	}

	err := utils.SetSessionUser(c, dbUser.ID)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("HX-Redirect", "/characters")
	c.Response().WriteHeader(300)

	return nil
}

func Logout(c echo.Context) error {
	utils.SetSessionUser(c, 0)

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func Register(c echo.Context) error {
	return templates.RenderGroup(c, http.StatusOK, "register.html", nil, authViewPatterns, registerViewPatterns)
}

func PostRegister(c echo.Context) error {
	userDto := new(models.User)

	if err := c.Bind(userDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	userDto.Password = utils.HashPassword(userDto.Password)

	user := models.InitUser(userDto)

	roleModel := models.NewRole()
	userRole := roleModel.GetRoleByName("User")
	if userRole.ID == 0 {
		return c.String(http.StatusInternalServerError, "Registration is not configured: the User role is missing")
	}

	// Set the foreign key explicitly; assigning the association made GORM upsert the
	// role row and left role_id at 0 when it did not.
	user.RoleId = userRole.ID

	_, err := user.Create()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save User")
	}

	c.Response().Header().Set("HX-Redirect", "/login")
	c.Response().WriteHeader(300)

	return nil
}

func ClearAlert(c echo.Context) error {
	return templates.RenderGroup(c, http.StatusOK, "oob-alert-empty", nil, authViewPatterns)
}
