package models

import (
	"errors"
	"fmt"
	"time"

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

// ErrActionInProgress means the action's runtime has not elapsed yet.
var ErrActionInProgress = errors.New("that action is still running")

// applyAction is one resolved click: bag space, ingredients, skill XP, character XP and
// the output item, all inside the caller's transaction.
func (model *Game) applyAction(tx *gorm.DB, character *Character, item *Item) (ActionResult, error) {
	var result ActionResult

	result.CharacterID = character.ID
	result.SkillID = uint(item.SkillId)
	result.Gained = *item

	// Bag space: an existing stack always accepts more; a new stack needs a free slot.
	var existing Inventory
	tx.Where("character_id = ? AND item_id = ? AND deleted_at IS NULL", character.ID, item.ID).First(&existing)

	if existing.ID == 0 {
		var stacks int64

		if err := tx.Raw(
			`select count(*) from inventories where character_id = ? and amount > 0 and deleted_at is null`,
			character.ID,
		).Scan(&stacks).Error; err != nil {
			return result, err
		}

		if stacks >= MaxInventorySize {
			return result, ErrInventoryFull
		}
	}

	// Ingredients: a conditional decrement closes the check-then-spend gap.
	var recipes []Recipe
	if err := tx.Where("parent_item_id = ?", item.ID).Find(&recipes).Error; err != nil {
		return result, err
	}

	for _, recipe := range recipes {
		res := tx.Exec(
			`UPDATE inventories SET amount = amount - ?
			 WHERE character_id = ? AND item_id = ? AND amount >= ? AND deleted_at IS NULL`,
			recipe.Amount, character.ID, recipe.ItemId, recipe.Amount,
		)
		if res.Error != nil {
			return result, res.Error
		}

		if res.RowsAffected == 0 {
			return result, ErrNotEnoughItems
		}

		tx.Exec(`DELETE FROM inventories WHERE character_id = ? AND item_id = ? AND amount <= 0`, character.ID, recipe.ItemId)
	}

	skillResult, err := model.applySkillXP(tx, character.ID, uint(item.SkillId), item.ExperienceGained)
	if err != nil {
		return result, err
	}

	result.SkillXP = skillResult.ExperiencePoints
	result.SkillLevel = skillResult.LevelName
	result.LeveledUp = skillResult.LeveledUp

	if err := model.applyCharacterXP(tx, character, item.ExperienceGained); err != nil {
		return result, err
	}

	// The output item, in the same transaction as the spend.
	if existing.ID != 0 {
		if err := tx.Exec(`UPDATE inventories SET amount = amount + 1 WHERE id = ?`, existing.ID).Error; err != nil {
			return result, err
		}
	} else {
		output := Inventory{CharacterId: character.ID, ItemId: item.ID, Amount: 1}
		if err := tx.Create(&output).Error; err != nil {
			return result, err
		}
	}

	return result, nil
}

// checkAffordable validates bag space and ingredients without changing anything, so an
// action that cannot be paid for is refused before its timer starts.
func (model *Game) checkAffordable(tx *gorm.DB, character *Character, item *Item) error {
	var existing Inventory
	tx.Where("character_id = ? AND item_id = ? AND deleted_at IS NULL", character.ID, item.ID).First(&existing)

	if existing.ID == 0 {
		var stacks int64
		if err := tx.Raw(
			`select count(*) from inventories where character_id = ? and amount > 0 and deleted_at is null`,
			character.ID,
		).Scan(&stacks).Error; err != nil {
			return err
		}

		if stacks >= MaxInventorySize {
			return ErrInventoryFull
		}
	}

	var recipes []Recipe
	if err := tx.Where("parent_item_id = ?", item.ID).Find(&recipes).Error; err != nil {
		return err
	}

	for _, recipe := range recipes {
		var have int64
		if err := tx.Raw(
			`select coalesce(sum(amount), 0) from inventories where character_id = ? and item_id = ? and deleted_at is null`,
			character.ID, recipe.ItemId,
		).Scan(&have).Error; err != nil {
			return err
		}

		if have < recipe.Amount {
			return ErrNotEnoughItems
		}
	}

	return nil
}

// StartAction arms an action: nothing is granted yet, the item's runtime is recorded and
// the output lands when CollectAction runs after it.
func (model *Game) StartAction(characterID uint, itemID uint) error {
	return model.db.Transaction(func(tx *gorm.DB) error {
		var character Character
		if err := tx.First(&character, "id = ?", characterID).Error; err != nil {
			return ErrNoCharacter
		}

		if character.ActionItemId != 0 {
			if time.Now().UnixMilli() < character.ActionReadyAt {
				return ErrActionInProgress
			}

			return nil // finished and waiting: this click collects it
		}

		var item Item
		if err := tx.First(&item, "id = ?", itemID).Error; err != nil {
			return fmt.Errorf("unknown item")
		}

		if err := model.checkAffordable(tx, &character, &item); err != nil {
			return err
		}

		ready := time.Now()
		if item.Runtime > 0 {
			ready = ready.Add(item.Runtime)
		}

		return tx.Model(&Character{}).Where("id = ?", character.ID).Updates(map[string]any{
			"action_item_id":  item.ID,
			"action_ready_at": ready.UnixMilli(),
		}).Error
	})
}

// CollectAction finishes the pending action once its runtime has elapsed.
func (model *Game) CollectAction(characterID uint) (ActionResult, error) {
	var result ActionResult

	err := model.db.Transaction(func(tx *gorm.DB) error {
		var character Character
		if err := tx.Preload("Level").First(&character, "id = ?", characterID).Error; err != nil {
			return ErrNoCharacter
		}

		if character.ActionItemId == 0 {
			return nil
		}

		if time.Now().UnixMilli() < character.ActionReadyAt {
			return ErrActionInProgress
		}

		pending := character.ActionItemId

		// Clear the pending action first: a failure below must not re-trigger forever.
		if err := tx.Model(&Character{}).Where("id = ?", character.ID).Updates(map[string]any{
			"action_item_id":  uint(0),
			"action_ready_at": int64(0),
		}).Error; err != nil {
			return err
		}

		var item Item
		if err := tx.Preload("Level").First(&item, "id = ?", pending).Error; err != nil {
			return fmt.Errorf("unknown item")
		}

		res, err := model.applyAction(tx, &character, &item)
		result = res

		return err
	})

	return result, err
}

// RunAction starts and collects in one transaction. Used by tests and by any caller that
// does not play through the UI; the HTTP flow uses StartAction/CollectAction.
func (model *Game) RunAction(characterID uint, itemID uint) (ActionResult, error) {
	var result ActionResult

	err := model.db.Transaction(func(tx *gorm.DB) error {
		var character Character
		if err := tx.Preload("Level").First(&character, "id = ?", characterID).Error; err != nil {
			return ErrNoCharacter
		}

		var item Item
		if err := tx.Preload("Level").First(&item, "id = ?", itemID).Error; err != nil {
			return fmt.Errorf("unknown item")
		}

		res, err := model.applyAction(tx, &character, &item)
		result = res

		return err
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
