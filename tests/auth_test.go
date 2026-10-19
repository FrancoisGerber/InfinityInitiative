package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"InfinityInitiative/data/models"
	"InfinityInitiative/routes"
	"InfinityInitiative/utils"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// stubRenderer stands in for the real template engine: the tests only care about
// status codes, but handlers call c.Render, so a renderer must exist.
type stubRenderer struct{}

func (stubRenderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	_, err := w.Write([]byte("rendered:" + name))
	return err
}

// signedSessionCookie builds a valid session cookie for a user id.
func signedSessionCookie(t *testing.T, store sessions.Store, userID uint) *http.Cookie {
	t.Helper()

	sess := sessions.NewSession(store, "session")
	sess.Values["userId"] = userID
	if err := sess.Save(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()); err != nil {
		t.Fatalf("could not save session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := sess.Save(req, rec); err != nil {
		t.Fatalf("could not save session: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie produced")
	}

	return cookies[0]
}

func newTestServer() (*echo.Echo, sessions.Store) {
	store := sessions.NewCookieStore([]byte("test-secret-that-is-long-enough-for-signing-32"))
	e := echo.New()
	e.Renderer = stubRenderer{}
	e.Use(middleware.Recover())
	e.Use(session.Middleware(store))

	// Echo hides handler errors behind a generic 500; print the cause so a failing
	// assertion shows what actually broke.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		fmt.Printf("handler error on %s %s: %+v\n", c.Request().Method, c.Request().URL.Path, err)
		e.DefaultHTTPErrorHandler(err, c)
	}

	routes.Router(e)

	return e, store
}

// A logged-in account without the admin role must not reach the content-management
// routes. Regression test for the bug that let any user edit items, users and roles.
func TestAdminRoutesRequireAdminRole(t *testing.T) {
	roleModel := models.NewRole()
	role := roleModel.GetRoleByName("User")
	if role.ID == 0 {
		t.Fatal("User role was not seeded")
	}

	user := models.InitUser(&models.User{
		Username: "pleb",
		Email:    "pleb@test.local",
		Password: utils.HashPassword("Passw0rd!123"),
		RoleId:   role.ID,
	})
	if _, err := user.Create(); err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	for _, path := range []string{"/users", "/roles", "/items", "/monsters", "/characters"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(cookie)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("%s as non-admin: got %d, want %d", path, rec.Code, http.StatusForbidden)
		}
	}
}

// The admin role still gets in.
func TestAdminRoutesAllowAdminRole(t *testing.T) {
	roleModel := models.NewRole()
	role := roleModel.GetRoleByName("Admin")
	if role.ID == 0 {
		t.Fatal("Admin role was not seeded")
	}

	user := models.InitUser(&models.User{
		Username: "boss",
		Email:    "boss@test.local",
		Password: utils.HashPassword("Passw0rd!123"),
		RoleId:   role.ID,
	})
	if _, err := user.Create(); err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("admin was refused: %d body=%q", rec.Code, rec.Body.String())
	}
}
