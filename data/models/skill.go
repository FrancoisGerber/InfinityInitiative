package models

import (
	"gorm.io/gorm"
)

type Skill struct {
	Base
	gorm.Model
	ID    uint   `json:"ID" form:"ID" query:"ID"`
	Name  string `json:"Name" form:"Name" query:"Name"`
	Image string `json:"Image" form:"Image" query:"Image"`
}

func NewSkill() Skill {
	model := Skill{}
	model.initDatabase()

	return model
}

func InitSkill(model *Skill) Skill {
	model.initDatabase()

	return *model
}

func (model *Skill) Migrate() error {
	migration := model.db.AutoMigrate(&model)

	for _, skill := range Skills {
		newSkill := model.GetSkillByName(skill)
		if newSkill.ID == 0 {
			newSkill := NewSkill()
			newSkill.Name = skill
			newSkill.Image = "/assets/img/skills/" + skill + ".png"

			newSkill.Create()
		}
	}

	return migration
}

func (model *Skill) GetAllSkills() (models []Skill) {
	model.db.Find(&models)

	for _, m := range models {
		m = InitSkill(&m)
	}

	return models
}

func (model *Skill) GetSkill(id int64) Skill {
	var skill Skill
	model.db.First(&skill, "id = ?", id)

	skill = InitSkill(&skill)

	return skill
}

func (model *Skill) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Skill) Update() (int64, error) {
	result := model.db.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Skill) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Skill) GetSkillByName(name string) Skill {
	var skill Skill
	model.db.First(&skill, "name = ?", name)

	skill = InitSkill(&skill)

	return skill
}
