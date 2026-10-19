package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrInventoryFull  = errors.New("your inventory is full")
	ErrNotEnoughItems = errors.New("you do not have the resources for this item")
	ErrNoCharacter    = errors.New("this account has no character yet")
)

// ActionResult describes one resolved click: what was made, what the skill looks like
// afterwards, and whether a level was gained.
type ActionResult struct {
	CharacterID uint
	SkillID     uint
	Gained      Item
	SkillXP     int64
	SkillLevel  int64
	LeveledUp   bool
}

// RunAction resolves one action atomically.
//
// The previous flow removed the ingredients, then slept inside the HTTP handler for up
// to a minute, then granted XP and the output — with no transaction, so a crash or a
// disconnect lost the ingredients while a duplicate request paid out twice.
func (model *Game) RunAction(userID uint, itemID uint) (ActionResult, error) {
	var result ActionResult

	err := model.db.Transaction(func(tx *gorm.DB) error {
		var character Character
		if err := tx.Preload("Level").First(&character, "user_id = ?", userID).Error; err != nil {
			return ErrNoCharacter
		}

		var item Item
		if err := tx.Preload("Level").First(&item, "id = ?", itemID).Error; err != nil {
			return fmt.Errorf("unknown item")
		}

		result.CharacterID = character.ID
		result.SkillID = uint(item.SkillId)
		result.Gained = item

		// Bag space: an existing stack always accepts more; a new stack needs a free slot.
		var existing Inventory
		tx.Where("character_id = ? AND item_id = ? AND deleted_at IS NULL", character.ID, item.ID).First(&existing)

		if existing.ID == 0 {
			var stacks int64

			err := tx.Raw(
				`select count(*) from inventories where character_id = ? and amount > 0 and deleted_at is null`,
				character.ID,
			).Scan(&stacks).Error
			if err != nil {
				return err
			}

			if stacks >= MaxInventorySize {
				return ErrInventoryFull
			}
		}

		// Ingredients: a conditional decrement closes the check-then-spend gap.
		var recipes []Recipe
		if err := tx.Where("parent_item_id = ?", item.ID).Find(&recipes).Error; err != nil {
			return err
		}

		for _, recipe := range recipes {
			res := tx.Exec(
				`UPDATE inventories SET amount = amount - ?
				 WHERE character_id = ? AND item_id = ? AND amount >= ? AND deleted_at IS NULL`,
				recipe.Amount, character.ID, recipe.ItemId, recipe.Amount,
			)
			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected == 0 {
				return ErrNotEnoughItems
			}

			tx.Exec(`DELETE FROM inventories WHERE character_id = ? AND item_id = ? AND amount <= 0`, character.ID, recipe.ItemId)
		}

		skillResult, err := model.applySkillXP(tx, character.ID, uint(item.SkillId), item.ExperienceGained)
		if err != nil {
			return err
		}

		result.SkillXP = skillResult.ExperiencePoints
		result.SkillLevel = skillResult.LevelName
		result.LeveledUp = skillResult.LeveledUp

		if err := model.applyCharacterXP(tx, &character, item.ExperienceGained); err != nil {
			return err
		}

		// The output item, in the same transaction as the spend.
		if existing.ID != 0 {
			if err := tx.Exec(`UPDATE inventories SET amount = amount + 1 WHERE id = ?`, existing.ID).Error; err != nil {
				return err
			}
		} else {
			output := Inventory{CharacterId: character.ID, ItemId: item.ID, Amount: 1}
			if err := tx.Create(&output).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

type skillProgress struct {
	ExperiencePoints int64
	LevelName        int64
	LeveledUp        bool
}

// applySkillXP adds XP and advances as many levels as the new total allows (the old code
// advanced at most one level per action).
func (model *Game) applySkillXP(tx *gorm.DB, characterID uint, skillID uint, gained int64) (skillProgress, error) {
	var progress skillProgress

	var characterSkill CharacterSkill
	err := tx.Where("character_id = ? AND skill_id = ?", characterID, skillID).First(&characterSkill).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var firstLevel Level
		tx.Where("name = ?", 1).First(&firstLevel)

		characterSkill = CharacterSkill{CharacterId: characterID, SkillId: skillID, LevelId: firstLevel.ID}
		if err := tx.Create(&characterSkill).Error; err != nil {
			return progress, err
		}
	} else if err != nil {
		return progress, err
	}

	characterSkill.ExperiencePoints += gained

	leveled := false
	for {
		var current Level
		if err := tx.First(&current, "id = ?", characterSkill.LevelId).Error; err != nil {
			return progress, err
		}

		var next Level
		if err := tx.Where("name = ?", current.Name+1).First(&next).Error; err != nil {
			break // at the level cap
		}

		if next.ExperienceNeeded <= 0 || characterSkill.ExperiencePoints < next.ExperienceNeeded {
			break
		}

		characterSkill.LevelId = next.ID
		leveled = true
	}

	if err := tx.Model(&CharacterSkill{}).Where("id = ?", characterSkill.ID).Updates(map[string]any{
		"experience_points": characterSkill.ExperiencePoints,
		"level_id":          characterSkill.LevelId,
	}).Error; err != nil {
		return progress, err
	}

	var finalLevel Level
	tx.First(&finalLevel, "id = ?", characterSkill.LevelId)

	progress.ExperiencePoints = characterSkill.ExperiencePoints
	progress.LevelName = finalLevel.Name
	progress.LeveledUp = leveled

	return progress, nil
}

// applyCharacterXP adds a fraction of the action's XP to the character itself. Character
// XP is its own counter now, instead of being recomputed as an average over every skill
// in the game.
func (model *Game) applyCharacterXP(tx *gorm.DB, character *Character, gained int64) error {
	character.ExperiencePoints += gained / 4

	for {
		var current Level
		if err := tx.First(&current, "id = ?", character.LevelId).Error; err != nil {
			return nil
		}

		var next Level
		if err := tx.Where("name = ?", current.Name+1).First(&next).Error; err != nil {
			break
		}

		if next.ExperienceNeeded <= 0 || character.ExperiencePoints < next.ExperienceNeeded {
			break
		}

		character.LevelId = next.ID
	}

	return tx.Model(&Character{}).Where("id = ?", character.ID).Updates(map[string]any{
		"experience_points": character.ExperiencePoints,
		"level_id":          character.LevelId,
	}).Error
}
