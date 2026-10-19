package main

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"InfinityInitiative/routes"
	"InfinityInitiative/utils"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	//Load environment variables
	err := godotenv.Load(".env")

	if err != nil {
		panic("Error loading .env file")
	}

	// The database must be named explicitly: an empty DBNAME made SQLite open its own
	// unnamed database, and a typo silently started the app on a brand-new empty file.
	if utils.DBName() == "" {
		panic("DBNAME must point at the SQLite file to use (see .env)")
	}

	// Session signing key: an empty key makes cookies forgeable, a short one is weak.
	if len(utils.Secret()) == 0 {
		panic("SECRET must be set (32+ random bytes), e.g. openssl rand -base64 48")
	}

	// Setup Echo
	e := echo.New()

	e.Logger.Infof("database: %s", utils.DBName())

	if len(utils.Secret()) < 32 {
		e.Logger.Warn("SECRET is shorter than 32 bytes — rotate it to a random 32+ byte value")
	}

	// Root level middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(utils.SessionStore()))

	// CSRF: every mutating request must carry the token. The cookie is readable by the
	// page's own script (HttpOnly off on purpose) so htmx can copy it into a header.
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token,form:_csrf",
		ContextKey:     "csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieHTTPOnly: false,
		CookieSecure:   !utils.InsecureCookies(),
		CookieSameSite: http.SameSiteLaxMode,
	}))

	// Static files and templates
	e.Static("/assets", "views/assets")

	// Router
	routes.Router(e)

	// Run DB migrations
	autoMigrations()

	// Start Server
	e.Logger.Fatal(e.Start(":8069"))
}

// Database auto migrations
func autoMigrations() {
	character := models.NewCharacter()
	characterSkill := models.NewCharacterSkill()
	level := models.NewLevel()
	item := models.NewItem()
	skill := models.NewSkill()
	recipe := models.NewRecipe()
	monster := models.NewMonster()
	inventory := models.NewInventory()
	role := models.NewRole()
	user := models.NewUser()

	data.AutoMigration(&character)
	data.AutoMigration(&characterSkill)
	data.AutoMigration(&level)
	data.AutoMigration(&item)
	data.AutoMigration(&skill)
	data.AutoMigration(&recipe)
	data.AutoMigration(&monster)
	data.AutoMigration(&inventory)
	data.AutoMigration(&role)
	data.AutoMigration(&user)
}
