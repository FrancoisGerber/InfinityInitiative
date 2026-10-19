package routes

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

// requireRole authorises an already-authenticated request against the user's role.
// It also loads the user row, so a deleted/demoted account loses access immediately
// instead of keeping it until the session cookie expires.
func requireRole(roleID uint) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("UserId").(uint)
			if !ok || userID == 0 {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			userModel := models.NewUser()
			user := userModel.GetUser(int64(userID))
			if user.ID == 0 {
				// Account no longer exists: drop the session and send them back to login.
				utils.SetSessionUser(c, 0)
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			if user.RoleId != roleID {
				if c.Request().Header.Get("HX-Request") == "true" {
					return c.String(http.StatusForbidden, "Forbidden — this area is for administrators.")
				}

				return c.HTML(http.StatusForbidden,
					`<!doctype html><html><body style="background:#0f192d;color:#e5e7eb;font-family:sans-serif;padding:4rem;text-align:center">`+
						`<h1 style="font-size:1.5rem;font-weight:700">403 — not your door</h1>`+
						`<p style="opacity:.7;margin-top:.5rem">This area is for administrators.</p>`+
						`<p style="margin-top:1.5rem"><a style="color:#7dd3fc" href="/characters">Back to your character</a></p>`+
						`</body></html>`)
			}

			c.Set("User", user)

			return next(c)
		}
	}
}
