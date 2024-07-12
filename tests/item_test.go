package tests

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"testing"

	"github.com/joho/godotenv"
)

func TestItem(t *testing.T) {
	err := godotenv.Load("../.env")

	if err != nil {
		panic("Error loading .env file")
	}

	item := models.NewItem()
	data.AutoMigration(&item)

	data := item.GetAllItemsForSkill(5)

	t.Log("DATA: ", data)
}
