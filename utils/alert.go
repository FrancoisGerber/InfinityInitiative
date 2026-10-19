package utils

import (
	"net/http"

	templates "InfinityInitiative/views"

	"github.com/labstack/echo/v4"
)

// The alert fragments live in the layout set, so they are rendered with it directly —
// the handlers no longer install a renderer on the shared Echo instance.
var alertViewPatterns = "views/layout/*.html"

func ShowAlert(c echo.Context, message string) {
	showAlert(c, message, "oob-alert")
}

func ShowAlertSuccess(c echo.Context, message string) {
	showAlert(c, message, "oob-alert-success")
}

func ShowAlertError(c echo.Context, message string) {
	showAlert(c, message, "oob-alert-error")
}

func showAlert(c echo.Context, message string, alertType string) {
	err := templates.RenderGroup(c, http.StatusOK, alertType, struct{ Message string }{message}, alertViewPatterns)
	if err != nil {
		c.Logger().Error(err)
	}
}
