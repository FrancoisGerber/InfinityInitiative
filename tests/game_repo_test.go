package tests

import (
	"InfinityInitiative/data/models"
	"testing"
)

func TestGameRepo(t *testing.T) {
	repo := models.NewGame("Test.db")

	repo.Log("Hallo from testing")
}
