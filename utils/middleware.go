package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SessionAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, err := GetSessionUser(c)

		if err != nil || userId == 0 {
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}

		c.Set("UserId", userId)

		// Let Echo's error handler deal with handler errors; swallowing them here made
		// every failure look like a 200.
		return next(c)
	}
}
