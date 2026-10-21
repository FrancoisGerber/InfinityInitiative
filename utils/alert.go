package utils

import (
	"net/http"

	templates "InfinityInitiative/views"

	"github.com/labstack/echo/v4"
)

// The alert fragments live in the layout set, so they are rendered with it directly —
// the handlers no longer install a renderer on the shared Echo instance.
var alertViewPatterns = "views/layout/*.html"

type alertData struct {
	Message string
}

// ShowAlertSuccess renders a success toast out of band, alongside a successful response.
func ShowAlertSuccess(c echo.Context, message string) error {
	return showOOB(c, "oob-alert-success", message)
}

// ShowAlertError answers a failed request with the error's real HTTP status and one alert
// fragment, retargeted at the toast container. The status matters: a refused login is a
// 401, a forbidden action a 403, a failed write a 500, not a 200 with a toast. htmx is
// configured (both shells) to swap 4xx/5xx responses, and HX-Retarget keeps the form that
// posted intact.
func ShowAlertError(c echo.Context, status int, message string) error {
	c.Response().Header().Set("HX-Retarget", "#alerts")
	c.Response().Header().Set("HX-Reswap", "innerHTML")

	return templates.RenderGroup(c, status, "alert-error", alertData{message}, alertViewPatterns)
}

func showOOB(c echo.Context, view string, message string) error {
	return templates.RenderGroup(c, http.StatusOK, view, alertData{message}, alertViewPatterns)
}
