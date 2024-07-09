package utils

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func GetSessionUser(c echo.Context) (uint, error) {
	sess, err := session.Get("session", c)

	if err != nil {
		return uint(0), err
	}

	userId := sess.Values["userId"]

	if userId == nil {
		userId = uint(0)
	}

	return userId.(uint), nil
}

func CreateSessionUser(c echo.Context, userId uint) error {
	sess, err := session.Get("session", c)

	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path: "/",
	}

	sess.Values["userId"] = uint(userId)

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
