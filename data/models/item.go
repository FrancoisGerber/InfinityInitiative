package models

import (
	"InfinityInitiative/data"
	"time"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	ID               uint          `json:"ID" form:"ID" query:"ID"`
	Name             string        `json:"Name" form:"Name" query:"Name"`
	Type             string        `json:"Type" form:"Type" query:"Type"`
	Description      string        `json:"Description" form:"Description" query:"Description"`
	Image            string        `json:"Image" form:"Image" query:"Image"`
	LevelId          int           `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level            Level         `json:"Level" form:"Level" query:"Level"`
	ExperienceGained int64         `json:"ExperienceGained" form:"ExperienceGained" query:"ExperienceGained"`
	Runtime          time.Duration `json:"Runtime" form:"Runtime" query:"Runtime"`
	RuntimeString    string        `json:"RuntimeString" form:"RuntimeString" query:"RuntimeString"`
}

func MigrateItems() {
	data.DB.AutoMigrate(&Item{})
}

func GetAllItems() (models []Item) {
	data.DB.Preload("Level").Find(&models)

	return models
}

func GetItem(id int64) (model Item) {
	data.DB.Preload("Level").First(&model, "id = ?", id)

	return model
}

func (model *Item) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Item) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Item) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
