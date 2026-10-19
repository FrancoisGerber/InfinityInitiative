package main

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"InfinityInitiative/routes"
	"InfinityInitiative/utils"
	"fmt"
	"net/http"
	"strings"

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

	if !utils.SecureCookies() {
		e.Logger.Warn("cookies are not marked Secure — set SECURE_COOKIES=1 when serving over HTTPS")
	}

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
		CookieSecure:   utils.SecureCookies(),
		CookieSameSite: http.SameSiteLaxMode,
	}))

	// Static files and templates
	e.Static("/assets", "views/assets")

	// Errors: a rejected CSRF token means the page the browser posted from is stale, which
	// is invisible as a bare 400.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		if he, ok := err.(*echo.HTTPError); ok && he.Code == http.StatusBadRequest &&
			(strings.Contains(fmt.Sprint(he.Message), "csrf") || he.Internal != nil && strings.Contains(fmt.Sprint(he.Internal), "csrf")) {
			c.Response().Header().Set("Cache-Control", "no-store")
			_ = c.HTML(http.StatusBadRequest, "<!doctype html><title>Session expired</title>"+
				"<body style=\"font-family:system-ui;padding:3rem;background:#0f192d;color:#e5e7eb\">"+
				"<h1>Session expired</h1><p>Your browser posted an out-of-date form (the security token was missing or stale).</p>"+
				"<p><a style=\"color:#93c5fd\" href=\"/login\">Reload the login page</a> and try again.</p></body>")
			return
		}

		e.DefaultHTTPErrorHandler(err, c)
	}

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
