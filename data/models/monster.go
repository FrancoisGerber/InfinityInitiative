package models

import (
	"gorm.io/gorm"
)

type Monster struct {
	Base
	gorm.Model
	ID               uint    `json:"ID" form:"ID" query:"ID"`
	Name             string  `json:"Name" form:"Name" query:"Name"`
	Description      string  `json:"Description" form:"Description" query:"Description"`
	Image            string  `json:"Image" form:"Image" query:"Image"`
	LevelId          uint    `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level            Level   `json:"Level" form:"Level" query:"Level"`
	ExperienceGained int64   `json:"ExperienceGained" form:"ExperienceGained" query:"ExperienceGained"`
	Health           int64   `json:"Health" form:"Health" query:"Health"`
	MaxHealth        int64   `json:"MaxHealth" form:"MaxHealth" query:"MaxHealth"`
	Attack           float64 `json:"Attack" form:"Attack" query:"Attack"`
	Strength         float64 `json:"Strength" form:"Strength" query:"Strength"`
	Defence          float64 `json:"Defence" form:"Defence" query:"Defence"`
	ArmourRating     float64 `json:"ArmourRating" form:"ArmourRating" query:"ArmourRating"`
}

func NewMonster() Monster {
	model := Monster{}
	model.initDatabase()

	return model
}

func InitMonster(model *Monster) Monster {
	model.initDatabase()

	return *model
}

func (model *Monster) Migrate() error {
	return model.db.AutoMigrate(&model)
}

func (model *Monster) GetAllMonsters() (models []Monster) {
	model.db.Preload("Level").Find(&models)

	for _, m := range models {
		m = InitMonster(&m)
	}

	return models
}

func (model *Monster) GetMonster(id int64) Monster {
	var monster Monster
	model.db.Preload("Level").First(&monster, "id = ?", id)

	monster = InitMonster(&monster)

	return monster
}

func (model *Monster) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Monster) Update() (int64, error) {
	result := model.db.Select("*").Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Monster) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
