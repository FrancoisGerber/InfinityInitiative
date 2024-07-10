package data

import (
	"InfinityInitiative/data/interfaces"

	"gorm.io/gorm"
)

// TODO: remove global DB once everything is in repos
var DB *gorm.DB

func Init() *gorm.DB {
	return InitLocal()
}

func AutoMigration(model interfaces.Model) error {
	err := model.Migrate()
	return err
}
