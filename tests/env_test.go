package tests

import (
	"InfinityInitiative/utils"
	"testing"

	"github.com/joho/godotenv"
)

func TestEnv(t *testing.T) {
	//Load environment variables
	err := godotenv.Load("../.env")

	if err != nil {
		panic("Error loading .env file")
	}

	dbName := utils.DBName()
	t.Log("db name is: ", dbName)

	if dbName == "" {
		t.Fail()
	}
}
