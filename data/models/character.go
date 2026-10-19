package models

import (
	"gorm.io/gorm"
)

type Character struct {
	Base
	gorm.Model
	ID                     uint    `json:"ID" form:"ID" query:"ID"`
	Name                   string  `json:"Name" form:"Name" query:"Name"`
	Class                  string  `json:"Class" form:"Class" query:"Class"`
	Bio                    string  `json:"Bio" form:"Bio" query:"Bio"`
	Image                  string  `json:"Image" form:"Image" query:"Image"`
	Gold                   float64 `json:"Gold" form:"Gold" query:"Gold" gorm:"default:100"`
	Diamonds               float64 `json:"Diamonds" form:"Diamonds" query:"Diamonds" gorm:"default:10"`
	Health                 int64   `json:"Health" form:"Health" query:"Health" gorm:"default:10"`
	MaxHealth              int64   `json:"MaxHealth" form:"MaxHealth" query:"MaxHealth" gorm:"default:10"`
	LevelId                uint    `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level                  Level   `json:"Level" form:"Level" query:"Level"`
	UserId                 uint    `json:"UserId" form:"UserId" query:"UserId"`
	User                   User    `json:"User" form:"User" query:"User"`
	Attack                 float64 `json:"Attack" form:"Attack" query:"Attack" gorm:"default:1"`
	Strength               float64 `json:"Strength" form:"Strength" query:"Strength" gorm:"default:1"`
	Defence                float64 `json:"Defence" form:"Defence" query:"Defence" gorm:"default:1"`
	ArmourRating           float64 `json:"ArmourRating" form:"ArmourRating" query:"ArmourRating" gorm:"default:0"`
	ExperiencePoints       int64   `json:"ExperiencePoints" form:"ExperiencePoints" query:"ExperiencePoints"`
	ActiveItemId           *uint   `json:"ActiveItemId" form:"ActiveItemId" query:"ActiveItemId"`
	ActiveItem             Item    `json:"ActiveItem" form:"ActiveItem" query:"ActiveItem"`
	Notes                  string  `json:"Notes" form:"Notes" query:"Notes"`
	CombatLevelId          uint    `json:"CombatLevelId" form:"CombatLevelId" query:"CombatLevelId"`
	CombatLevel            Level   `json:"CombatLevel" form:"CombatLevel" query:"CombatLevel"`
	CombatExperiencePoints int64   `json:"CombatExperiencePoints" form:"CombatExperiencePoints" query:"CombatExperiencePoints"`
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
	var character Character
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&character, "id = ?", id)

	character = InitCharacter(&character)

	return character
}

func (model *Character) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	for _, s := range Skills {
		skill := NewSkill()
		skill = skill.GetSkillByName(s)

		characterSkill := NewCharacterSkill()
		characterSkill.GetCharacterProgress(model.ID, skill.ID)
	}

	return result.RowsAffected, nil
}

func (model *Character) Update() (int64, error) {
	result := updateAll(model.db, model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) Delete() (int64, error) {
	result := model.db.Delete(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Character) GetCharacterForUser(userId int64) Character {
	var character Character
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&character, "user_id = ?", userId)

	character = InitCharacter(&character)

	return character
}

func (model *Character) GetFullCharacterForUser(userId int64) FullCharacter {
	var character Character
	model.db.Preload("User").Preload("Level").Preload("ActiveItem").First(&character, "user_id = ?", userId)

	character = InitCharacter(&character)
	characterSkill := NewCharacterSkill()
	characterSkills := characterSkill.GetAllCharacterSkillsPerCharacter(int64(character.ID))

	character.LoadCharacterActualStats(int64(character.ID))

	fullCharacter := FullCharacter{
		character,
		characterSkills,
	}

	return fullCharacter
}

func (model *Character) LoadCharacterActualStats(characterId int64) {
	inventory := NewInventory()
	inventoryList := inventory.GetInventoryForCharacter(int64(model.ID))
	for _, inventoryItem := range inventoryList {
		if inventoryItem.Equipped {
			model.Attack += inventoryItem.Item.Attack
			model.Strength += inventoryItem.Item.Strength
			model.Defence += inventoryItem.Item.Defence
			model.ArmourRating += inventoryItem.Item.ArmourRating
		}
	}
}
