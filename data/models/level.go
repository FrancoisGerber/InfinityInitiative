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

	for levelName := range 100 {
		level := GetLevelByName(int64(levelName))
		if level.ID == 0 && levelName > 0 {
			data.DB.Create(&Level{
				Name:             int64(levelName),
				ExperienceNeeded: int64(((levelName - 1) * 1000) * levelName),
			})
		}
	}
}

func GetAllLevels() (models []Level) {
	data.DB.Find(&models)

	return models
}

func GetLevel(id int64) (model Level) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func GetLevelByName(name int64) (model Level) {
	data.DB.First(&model, "name = ?", name)

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
