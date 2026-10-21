package auth

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"InfinityInitiative/views"
	"github.com/labstack/echo/v4"
	"net/http"
	"slices"
	"strings"
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
		return utils.ShowAlertError(c, http.StatusBadRequest, "Could not read the login form. Reload the page and try again.")
	}

	user := models.InitUser(userDto)

	dbUser := user.GetUserByEmail(user.Email)

	if dbUser.ID == 0 || dbUser.Password == "" {
		return utils.ShowAlertError(c, http.StatusUnauthorized, "Email not registered!")
	}

	verified := utils.VerifyPassword(user.Password, dbUser.Password)

	if !verified {
		return utils.ShowAlertError(c, http.StatusUnauthorized, "Password is incorrect!")
	}

	err := utils.SetSessionUser(c, dbUser.ID)

	if err != nil {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could not start a session. Try again.")
	}

	// "/" resolves to the player's first skill. Redirecting to /characters sent every
	// non-admin to the admin CRUD screen, which answers with the 403 page.
	c.Response().Header().Set("HX-Redirect", "/")
	c.Response().WriteHeader(300)

	return nil
}

func Logout(c echo.Context) error {
	utils.SetSessionUser(c, 0)

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func Register(c echo.Context) error {
	return templates.RenderGroup(c, http.StatusOK, "register.html", models.Data{
		DropdownData: map[string]any{
			"Classes":    models.Classes,
			"Characters": 0,
		},
	}, authViewPatterns, registerViewPatterns)
}

// registrationDto carries the account *and* the first character: an account without a
// character could not play at all, and the game page used to write rows for character 0.
type registrationDto struct {
	Email         string
	Username      string
	Password      string
	CharacterName string
	Class         string
}

func PostRegister(c echo.Context) error {
	// Read the fields explicitly: Echo's binder needs `form` tags on the struct, and this
	// DTO is not a model (a missing tag silently bound every field as empty).
	dto := registrationDto{
		Email:         c.FormValue("Email"),
		Username:      c.FormValue("Username"),
		Password:      c.FormValue("Password"),
		CharacterName: c.FormValue("CharacterName"),
		Class:         c.FormValue("Class"),
	}

	if len(dto.Password) < 8 {
		return utils.ShowAlertError(c, http.StatusBadRequest, "Password must be at least 8 characters.")
	}

	characterName := strings.TrimSpace(dto.CharacterName)
	if characterName == "" {
		characterName = dto.Username
	}

	if len(characterName) < 3 || len(characterName) > 24 {
		return utils.ShowAlertError(c, http.StatusBadRequest, "Character name must be between 3 and 24 characters.")
	}

	if !slices.Contains(models.Classes, dto.Class) {
		dto.Class = models.Classes[0]
	}

	roleModel := models.NewRole()
	userRole := roleModel.GetRoleByName("User")
	if userRole.ID == 0 {
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Registration is not configured: the User role is missing")
	}

	user := models.InitUser(&models.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: utils.HashPassword(dto.Password),
		// Set the foreign key explicitly; assigning the association made GORM upsert the
		// role row and left role_id at 0 when it did not.
		RoleId: userRole.ID,
	})

	if _, err := user.Create(); err != nil {
		return utils.ShowAlertError(c, http.StatusConflict, "That email or username is already taken.")
	}

	level := models.NewLevel()
	firstLevel := level.GetLevelByName(1)

	character := models.InitCharacter(&models.Character{
		Name:          characterName,
		Class:         dto.Class,
		UserId:        user.ID,
		LevelId:       firstLevel.ID,
		CombatLevelId: firstLevel.ID,
		Health:        10,
		MaxHealth:     10,
		Attack:        1,
		Strength:      1,
		Defence:       1,
	})

	if _, err := character.Create(); err != nil {
		// The account would be unusable without a character, so do not keep it.
		user.Delete()
		return utils.ShowAlertError(c, http.StatusInternalServerError, "Could not create the character, please try again.")
	}

	c.Response().Header().Set("HX-Redirect", "/login")
	c.Response().WriteHeader(300)

	return nil
}

// ClearAlert empties the alert container after its toast has been shown. It must not
// remove the container itself: htmx could then never target it again, and every later
// alert on that page would silently disappear.
func ClearAlert(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
