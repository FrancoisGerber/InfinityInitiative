package models

import (
	"gorm.io/gorm"
	"time"
)

type Item struct {
	Base
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
	SkillId          int           `json:"SkillId" form:"SkillId" query:"SkillId"`
	Skill            Skill         `json:"Skill" form:"Skill" query:"Skill"`
}

func NewItem() Item {
	model := Item{}
	model.initDatabase()

	return model
}

func InitItem(model *Item) Item {
	model.initDatabase()

	return *model
}

func (model *Item) Migrate() error {
	return model.db.AutoMigrate(&model)
}

func (model *Item) GetAllItems() (models []Item) {
	model.db.Preload("Level").Preload("Skill").Find(&models)

	for _, m := range models {
		m = InitItem(&m)
	}

	return models
}

func (model *Item) GetAllItemsForSkill(skillId uint) (models []Item) {
	model.db.Preload("Level").Preload("Skill").Find(&models, "skill_id = ?", skillId)

	for _, m := range models {
		m = InitItem(&m)
	}

	return models
}

func (model *Item) GetItem(id int64) Item {
	var item Item
	model.db.Preload("Level").Preload("Skill").First(&item, "id = ?", id)

	item = InitItem(&item)

	return item
}

func (model *Item) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Item) Update() (int64, error) {
	result := model.db.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Item) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
