package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type Level struct {
	gorm.Model
	ID               uint  `json:"ID" form:"ID" query:"ID"`
	Name             int64 `json:"Name" form:"Name" query:"Name"`
	ExperienceNeeded int64 `json:"ExperienceNeeded" form:"ExperienceNeeded" query:"ExperienceNeeded"`
}

func MigrateLevels() {
	data.DB.AutoMigrate(&Level{})
}

func GetAllLevels() (models []Level) {
	data.DB.Find(&models)

	return models
}

func GetLevel(id int64) (model Level) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func (model *Level) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Level) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Level) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
