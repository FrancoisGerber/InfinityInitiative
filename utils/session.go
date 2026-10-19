package utils

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func sessionOptions() *sessions.Options {
	// Secure is on whenever the app is served over TLS (production/ngrok) and off for
	// plain-http local development, which would otherwise never receive the cookie.
	secure := os.Getenv("INSECURE_COOKIES") != "1"

	return &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	}
}

func GetSessionUser(c echo.Context) (uint, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return uint(0), err
	}

	userId := sess.Values["userId"]
	if userId == nil {
		userId = uint(0)
	}

	value, ok := userId.(uint)
	if !ok {
		return uint(0), nil
	}

	return value, nil
}

func SetSessionUser(c echo.Context, userId uint) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Options = sessionOptions()

	sess.Values["userId"] = uint(userId)
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

func GetGameState(c echo.Context) GameState {
	sess, err := session.Get("session", c)
	if err != nil {
		return GameState{}
	}

	gameStateValue := sess.Values["GameState"]
	if gameStateValue != nil {
		gameState := GameState{}
		json.Unmarshal([]byte(gameStateValue.(string)), &gameState)

		return gameState
	}

	return GameState{}
}

func SetGameState(c echo.Context, gameState GameState) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Options = sessionOptions()

	json, _ := json.Marshal(gameState)

	sess.Values["GameState"] = string(json)
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
