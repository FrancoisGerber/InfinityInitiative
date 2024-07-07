package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type Character struct {
	gorm.Model
	ID               uint    `json:"ID" form:"ID" query:"ID"`
	Name             string  `json:"Name" form:"Name" query:"Name"`
	Class            string  `json:"Class" form:"Class" query:"Class"`
	Bio              string  `json:"Bio" form:"Bio" query:"Bio"`
	Image            string  `json:"Image" form:"Image" query:"Image"`
	Gold             float64 `json:"Gold" form:"Gold" query:"Gold"`
	Diamonds         float64 `json:"Diamonds" form:"Diamonds" query:"Diamonds"`
	Health           int64   `json:"Health" form:"Health" query:"Health"`
	MaxHealth        int64   `json:"MaxHealth" form:"MaxHealth" query:"MaxHealth"`
	LevelId          uint    `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level            Level   `json:"Level" form:"Level" query:"Level"`
	UserId           uint    `json:"UserId" form:"UserId" query:"UserId"`
	User             User    `json:"User" form:"User" query:"User"`
	Attack           float64 `json:"Attack" form:"Attack" query:"Attack"`
	Strength         float64 `json:"Strength" form:"Strength" query:"Strength"`
	Defence          float64 `json:"Defence" form:"Defence" query:"Defence"`
	ArmourRating     float64 `json:"ArmourRating" form:"ArmourRating" query:"ArmourRating"`
	ExperiencePoints int64   `json:"ExperiencePoints" form:"ExperiencePoints" query:"ExperiencePoints"`
}

func MigrateCharacters() {
	data.DB.AutoMigrate(&Character{})
}

func GetAllCharacters() (models []Character) {
	data.DB.Preload("User").Preload("Level").Find(&models)

	return models
}

func GetCharacter(id int64) (model Character) {
	data.DB.Preload("User").Preload("Level").First(&model, "id = ?", id)

	return model
}

func (model *Character) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
