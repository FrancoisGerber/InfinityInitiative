package models

import (
	"gorm.io/gorm"
)

type Character struct {
	Base
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
	ActiveItemId     *uint   `json:"ActiveItemId" form:"ActiveItemId" query:"ActiveItemId"`
	ActiveItem       Item    `json:"ActiveItem" form:"ActiveItem" query:"ActiveItem"`
}

type FullCharacter struct {
	Character
	Skills []CharacterSkill
}

func NewCharacter() Character {
	model := Character{}
	model.initDatabase()

	return model
}

func InitCharacter(model *Character) Character {
	model.initDatabase()

	return *model
}

func (model *Character) Migrate() error {
	return model.db.AutoMigrate(&model)
}

func (model *Character) GetAllCharacters() (models []Character) {
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").Find(&models)

	for _, m := range models {
		m = InitCharacter(&m)
	}

	return models
}

func (model *Character) GetCharacter(id int64) Character {
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&model, "id = ?", id)

	return *model
}

func (model *Character) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	//TODO: Create character skills for every skill

	return result.RowsAffected, nil
}

func (model *Character) Update() (int64, error) {
	result := model.db.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) GetCharacterForUser(userId int64) Character {
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&model, "user_id = ?", userId)

	return *model
}

func (model *Character) GetFullCharacterForUser(userId int64) FullCharacter {
	var character Character
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&character, "user_id = ?", userId)

	characterSkill := NewCharacterSkill()
	characterSkills := characterSkill.GetAllCharacterSkillsPerCharacter(int64(character.ID))

	fullCharacter := FullCharacter{
		character,
		characterSkills,
	}

	return fullCharacter
}
