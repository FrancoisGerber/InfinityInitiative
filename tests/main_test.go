package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"InfinityInitiative/data/models"
)

// TestMain gives the whole package a throwaway database, so a test run never touches
// the developer's real data and every run starts from the same state (the old suite
// only passed the first time because it reused a gitignored fixture DB).
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ii-tests-")
	if err != nil {
		fmt.Println("could not create temp dir:", err)
		os.Exit(1)
	}

	os.Setenv("DBNAME", filepath.Join(dir, "test.db"))

	// Handlers resolve templates relative to the process working directory ("views/..."),
	// exactly as the server does when started from the repo root. The test binary starts
	// in tests/, so step up once.
	if err := os.Chdir(".."); err != nil {
		fmt.Println("could not chdir to the repo root:", err)
		os.Exit(1)
	}

	// Minimal schema for the models the tests touch.
	user := models.NewUser()
	role := models.NewRole()
	character := models.NewCharacter()
	level := models.NewLevel()
	skill := models.NewSkill()
	characterSkill := models.NewCharacterSkill()
	item := models.NewItem()
	recipe := models.NewRecipe()
	inventory := models.NewInventory()

	for _, m := range []interface{ Migrate() error }{
		&level, &skill, &role, &user, &character,
		&characterSkill, &item, &recipe, &inventory,
	} {
		if err := m.Migrate(); err != nil {
			fmt.Println("migration failed:", err)
			os.Exit(1)
		}
	}

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}
