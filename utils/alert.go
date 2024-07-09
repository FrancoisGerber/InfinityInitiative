package utils

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

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
	c.Render(http.StatusOK, alertType, struct{ Message string }{message})
}
