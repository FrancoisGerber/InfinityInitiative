package models

import (
	"gorm.io/gorm"
)

type Level struct {
	Base
	gorm.Model
	ID               uint  `json:"ID" form:"ID" query:"ID"`
	Name             int64 `json:"Name" form:"Name" query:"Name"`
	ExperienceNeeded int64 `json:"ExperienceNeeded" form:"ExperienceNeeded" query:"ExperienceNeeded"`
}

func NewLevel() Level {
	model := Level{}
	model.initDatabase()

	return model
}

func InitLevel(model *Level) Level {
	model.initDatabase()

	return *model
}

func (model *Level) Migrate() error {
	migration := model.db.AutoMigrate(&model)

	for levelName := range 100 {
		level := model.GetLevelByName(int64(levelName))
		if level.ID == 0 && levelName > 0 {
			model.db.Create(&Level{
				Name:             int64(levelName),
				ExperienceNeeded: int64(((levelName - 1) * 1000) * levelName),
			})
		}
	}

	return migration
}

func (model *Level) GetAllLevels() (models []Level) {
	model.db.Find(&models)

	for _, m := range models {
		m = InitLevel(&m)
	}

	return models
}

func (model *Level) GetLevel(id int64) Level {
	var level Level
	model.db.First(&level, "id = ?", id)

	level = InitLevel(&level)

	return level
}

func (model *Level) GetLevelByName(name int64) Level {
	var level Level
	model.db.First(&level, "name = ?", name)

	level = InitLevel(&level)

	return level
}

func (model *Level) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Level) Update() (int64, error) {
	result := model.db.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Level) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
