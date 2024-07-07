package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type CharacterSkill struct {
	gorm.Model
	ID               uint      `json:"ID" form:"ID" query:"ID"`
	CharacterId      uint      `json:"CharacterId" form:"CharacterId" query:"CharacterId"`
	Character        Character `json:"Character" form:"Character" query:"Character"`
	SkillId          uint      `json:"SkillId" form:"SkillId" query:"SkillId"`
	Skill            Skill     `json:"Skill" form:"Skill" query:"Skill"`
	LevelId          uint      `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level            Level     `json:"Level" form:"Level" query:"Level"`
	ExperiencePoints int64     `json:"ExperiencePoints" form:"ExperiencePoints" query:"ExperiencePoints"`
}

func MigrateCharacterSkills() {
	data.DB.AutoMigrate(&CharacterSkill{})
}

func GetAllCharacterSkills() (models []CharacterSkill) {
	data.DB.Find(&models)

	return models
}

func GetCharacterSkill(id int64) (model CharacterSkill) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func (model *CharacterSkill) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *CharacterSkill) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *CharacterSkill) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
