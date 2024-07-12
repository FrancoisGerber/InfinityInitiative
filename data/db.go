package data

import (
	"InfinityInitiative/data/interfaces"

	"gorm.io/gorm"
)

func Init() *gorm.DB {
	return InitLocal()
}

func AutoMigration(model interfaces.Model) error {
	err := model.Migrate()
	return err
}
