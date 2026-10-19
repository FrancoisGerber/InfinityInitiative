package utils

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// SecureCookies reports whether cookie traffic is TLS.
//
// The default is OFF: a Secure cookie sent over plain http:// is dropped by the browser
// (even on localhost), which silently breaks both the session and the CSRF double-submit
// cookie — every POST then fails with 400. Set SECURE_COOKIES=1 whenever the app is
// actually served over HTTPS (ngrok, a proxy, a real deployment).
func SecureCookies() bool {
	if os.Getenv("INSECURE_COOKIES") == "1" {
		return false
	}

	return os.Getenv("SECURE_COOKIES") == "1"
}

func sessionOptions() *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   SecureCookies(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30, // 30 days
	}
}

// SessionStore keeps the session on the server and only a signed id in the cookie.
//
// The cookie store this replaces serialised the whole session — including the GameState
// JSON — into a ~4 KB cookie, so a few more UI flags would have silently dropped the
// session. Stored under the user cache dir, outside the repo.
func SessionStore() sessions.Store {
	dir := os.Getenv("SESSION_DIR")
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			base = os.TempDir()
		}

		dir = filepath.Join(base, "infinity-initiative", "sessions")
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		base := filepath.Join(os.TempDir(), "infinity-initiative-sessions")
		if mkErr := os.MkdirAll(base, 0o700); mkErr == nil {
			dir = base
		}
	}

	// securecookie wants a 16/24/32-byte block key; SECRET is arbitrary length, so derive
	// a fixed-size key instead of trusting its length.
	key := sha256.Sum256([]byte(Secret()))

	store := sessions.NewFilesystemStore(dir, key[:])
	store.Options = sessionOptions()

	return store
}

// sessionUserID normalises the stored id: the value is a uint written by this package, but
// a different store or an older cookie could hand back an int.
func sessionUserID(value any) uint {
	switch v := value.(type) {
	case uint:
		return v
	case int:
		if v > 0 {
			return uint(v)
		}
	case int64:
		if v > 0 {
			return uint(v)
		}
	case float64:
		if v > 0 {
			return uint(v)
		}
	}

	return 0
}

func GetSessionUser(c echo.Context) (uint, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return uint(0), err
	}

	return sessionUserID(sess.Values["userId"]), nil
}

// SetSessionUser logs the user in and rotates the session id, so a session fixed before
// login (an attacker-supplied id) cannot be reused afterwards.
func SetSessionUser(c echo.Context, userId uint) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Options = sessionOptions()
	sess.ID = ""
	sess.Values["userId"] = uint(userId)

	return sess.Save(c.Request(), c.Response())
}

func GetGameState(c echo.Context) GameState {
	sess, err := session.Get("session", c)
	if err != nil {
		return GameState{}
	}

	gameStateValue := sess.Values["GameState"]
	if gameStateValue != nil {
		state, ok := gameStateValue.(string)
		if !ok {
			return GameState{}
		}

		gameState := GameState{}
		if err := json.Unmarshal([]byte(state), &gameState); err != nil {
			return GameState{}
		}

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

	state, err := json.Marshal(gameState)
	if err != nil {
		return err
	}

	sess.Values["GameState"] = string(state)

	return sess.Save(c.Request(), c.Response())
}

// GetSessionCharacter returns the character the player is currently controlling (0 when
// nothing is selected yet — an account may own several).
func GetSessionCharacter(c echo.Context) uint {
	sess, err := session.Get("session", c)
	if err != nil {
		return 0
	}

	return sessionUserID(sess.Values["characterId"])
}

// SetSessionCharacter pins the active character. Ownership must be checked by the caller.
func SetSessionCharacter(c echo.Context, characterId uint) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Options = sessionOptions()
	sess.Values["characterId"] = uint(characterId)

	return sess.Save(c.Request(), c.Response())
}
