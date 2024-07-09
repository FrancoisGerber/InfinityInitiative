package data

import (
	"InfinityInitiative/data/interfaces"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TODO: remove global DB once everything is in repos
var DB *gorm.DB

func Init(dbName string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("Failed to connect database")
	}

	DB = db
	return db
}

func AutoMigration(model interfaces.Model) error {
	err := model.Migrate()
	return err
}
