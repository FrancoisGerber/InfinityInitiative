package data

import (
	"InfinityInitiative/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitLocal() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(utils.DBName()), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}

	DB = db
	return db
}
