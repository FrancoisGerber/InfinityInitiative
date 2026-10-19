package models

import (
	"gorm.io/gorm"
)

type CharacterSkill struct {
	Base
	gorm.Model
	ID               uint      `json:"ID" form:"ID" query:"ID"`
	CharacterId      uint      `json:"CharacterId" form:"CharacterId" query:"CharacterId" gorm:"uniqueIndex:idx_character_skill"`
	Character        Character `json:"Character" form:"Character" query:"Character"`
	SkillId          uint      `json:"SkillId" form:"SkillId" query:"SkillId" gorm:"uniqueIndex:idx_character_skill"`
	Skill            Skill     `json:"Skill" form:"Skill" query:"Skill"`
	LevelId          uint      `json:"LevelId" form:"LevelId" query:"LevelId"`
	Level            Level     `json:"Level" form:"Level" query:"Level"`
	ExperiencePoints int64     `json:"ExperiencePoints" form:"ExperiencePoints" query:"ExperiencePoints"`
}

func NewCharacterSkill() CharacterSkill {
	model := CharacterSkill{}
	model.initDatabase()

	return model
}

func InitCharacterSkill(model *CharacterSkill) CharacterSkill {
	model.initDatabase()

	return *model
}

func (model *CharacterSkill) Migrate() error {
	return model.db.AutoMigrate(&model)
}

func (model *CharacterSkill) GetAllCharacterSkills() (models []CharacterSkill) {
	model.db.Preload("Level").Preload("Character").Preload("Skill").Find(&models)

	for _, m := range models {
		m = InitCharacterSkill(&m)
	}

	return models
}

func (model *CharacterSkill) GetCharacterSkill(id int64) CharacterSkill {
	var characterSkill CharacterSkill
	model.db.Preload("Level").Preload("Character").Preload("Skill").First(&characterSkill, "id = ?", id)

	characterSkill = InitCharacterSkill(&characterSkill)

	return characterSkill
}

func (model *CharacterSkill) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *CharacterSkill) Update() (int64, error) {
	result := updateAll(model.db, model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *CharacterSkill) Delete() (int64, error) {
	result := model.db.Delete(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *CharacterSkill) GetCharacterProgress(characterId uint, skillId uint) CharacterSkill {
	level1 := NewLevel()
	level1 = level1.GetLevelByName(1)

	// Get-or-create. The unique index (character_id, skill_id) makes the concurrent
	// double-insert impossible; FirstOrCreate keeps the happy path a single statement.
	seed := InitCharacterSkill(&CharacterSkill{
		CharacterId: characterId,
		SkillId:     skillId,
		LevelId:     level1.ID,
	})

	model.db.Where("character_id = ? AND skill_id = ?", characterId, skillId).FirstOrCreate(&seed)

	var characterSkill CharacterSkill
	model.db.Preload("Level").Preload("Character").Preload("Skill").
		Where("character_id = ? AND skill_id = ?", characterId, skillId).
		First(&characterSkill)

	return InitCharacterSkill(&characterSkill)
}

func (model *CharacterSkill) CalcCharacterTotalXP(characterId uint) int64 {
	var TotalExperiencePoints int64 = 0

	model.db.Raw(`
		select sum(experience_points) as TotalExperiencePoints
		from character_skills
		where character_id = ?
	`, characterId).Scan(&TotalExperiencePoints)

	return TotalExperiencePoints
}

func (model *CharacterSkill) GetAllCharacterSkillsPerCharacter(characterId int64) (models []CharacterSkill) {
	model.db.Preload("Level").Preload("Character").Preload("Skill").Find(&models, "character_id = ?", characterId)

	for _, m := range models {
		m = InitCharacterSkill(&m)
	}

	return models
}
