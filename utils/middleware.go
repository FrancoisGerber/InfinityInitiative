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

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
