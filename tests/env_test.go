package tests

import (
	"strings"
	"testing"

	"InfinityInitiative/utils"

	"github.com/joho/godotenv"
)

// The suite must always talk to a throwaway database — a stray run once wrote a test
// account into the developer's real data.
func TestEnv(t *testing.T) {
	// Working directory is the repo root (see TestMain in main_test.go).
	_ = godotenv.Load(".env")

	dbName := utils.DBName()
	if dbName == "" {
		t.Fatal("DBNAME is empty")
	}

	if !strings.Contains(dbName, "ii-tests-") {
		t.Fatalf("tests are not pointed at a temp database: DBNAME=%s", dbName)
	}

	t.Log("db name is:", dbName)
}

// A Secure cookie is dropped by the browser over plain http:// (localhost included),
// which silently kills the session *and* the CSRF double-submit cookie: every login POST
// then fails with 400. The flag must stay off unless explicitly enabled.
func TestCookieSecureFlag(t *testing.T) {
	t.Setenv("SECURE_COOKIES", "")
	t.Setenv("INSECURE_COOKIES", "")

	if utils.SecureCookies() {
		t.Fatal("cookies are marked Secure by default: plain-http logins will fail")
	}

	t.Setenv("SECURE_COOKIES", "1")
	if !utils.SecureCookies() {
		t.Fatal("SECURE_COOKIES=1 was ignored")
	}

	t.Setenv("INSECURE_COOKIES", "1")
	if utils.SecureCookies() {
		t.Fatal("INSECURE_COOKIES=1 must win over SECURE_COOKIES")
	}
}
