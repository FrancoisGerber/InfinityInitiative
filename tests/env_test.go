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
