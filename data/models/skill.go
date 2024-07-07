package models

import (
	"InfinityInitiative/data"
	"time"

	"gorm.io/gorm"
)

type Skill struct {
	gorm.Model
	ID               uint          `json:"ID" form:"ID" query:"ID"`
	Name             string        `json:"Name" form:"Name" query:"Name"`
	Image            string        `json:"Image" form:"Image" query:"Image"`
	ExperienceGained int64         `json:"ExperienceGained" form:"ExperienceGained" query:"ExperienceGained"`
	Runtime          time.Duration `json:"Runtime" form:"Runtime" query:"Runtime"`
	RuntimeString    string        `json:"RuntimeString" form:"RuntimeString" query:"RuntimeString"`
}

func MigrateSkills() {
	data.DB.AutoMigrate(&Skill{})
}

func GetAllSkills() (models []Skill) {
	data.DB.Find(&models)

	return models
}

func GetSkill(id int64) (model Skill) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func (model *Skill) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Skill) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Skill) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func GetSkillByName(name string) (model Skill) {
	data.DB.First(&model, "name = ?", name)

	return model
}
