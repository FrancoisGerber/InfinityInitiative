package data

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbName string = "InfinityInitiative.db"
var DB *gorm.DB

func Init() {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}

	DB = db
}
