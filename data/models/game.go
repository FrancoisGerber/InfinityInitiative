package models

import (
	"slices"
	"time"
)

type Game struct {
	Base
}

func NewGame() Game {
	model := Game{}
	model.initDatabase()

	return model
}

// Get complete character skill progress
func (model *Game) LoadCharacterProgress(character Character, skillId uint) struct {
	CharacterSkill
	NextLevel    Level
	ActiveItem   Item
	PercentageXP float64
	// Pacing of the action that is currently running (the runtime shown in the UI).
	Working     bool
	WaitMs      int64
	SecondsLeft int64
} {

	characterSkill := NewCharacterSkill()
	level := NewLevel()
	characterSkill = characterSkill.GetCharacterProgress(character.ID, skillId)

	nextLevelName := characterSkill.Level.Name + 1
	nextLevel := level.GetLevelByName(nextLevelName)

	var activeItem Item
	if character.ActiveItem.SkillId == int(skillId) {
		activeItem = character.ActiveItem
	}

	newMax := nextLevel.ExperienceNeeded - characterSkill.Level.ExperienceNeeded
	newXp := characterSkill.ExperiencePoints - characterSkill.Level.ExperienceNeeded

	// At the level cap there is no "next" level, so the divisor collapses to zero. Guard
	// it instead of rendering NaN/Inf into the XP bar.
	percentageXP := 0.0
	if newMax > 0 {
		percentageXP = float64(newXp) / float64(newMax) * 100
		if percentageXP < 0 {
			percentageXP = 0
		}
		if percentageXP > 100 {
			percentageXP = 100
		}
	}

	now := time.Now().UnixMilli()
	waitMs := int64(0)
	working := character.ActionItemId != 0 && character.ActionReadyAt > now
	if working {
		waitMs = character.ActionReadyAt - now
	}

	characterSkillDto := struct {
		CharacterSkill
		NextLevel    Level
		ActiveItem   Item
		PercentageXP float64
		Working      bool
		WaitMs       int64
		SecondsLeft  int64
	}{
		characterSkill,
		nextLevel,
		activeItem,
		percentageXP,
		working,
		waitMs,
		(waitMs + 999) / 1000,
	}

	return characterSkillDto
}

// Get complete Items with recipes
func (model *Game) LoadFullItems(skill *Skill, character Character) []struct {
	Item
	Recipes    []Recipe
	ActiveItem Item
} {

	item := NewItem()
	items := item.GetAllItemsForSkill(skill.ID)

	var itemDtos []struct {
		Item
		Recipes    []Recipe
		ActiveItem Item
	}

	for _, item := range items {
		recipe := NewRecipe()
		recipes := recipe.GetRecipesForParentItem(int64(item.ID))

		itemDto := struct {
			Item
			Recipes    []Recipe
			ActiveItem Item
		}{
			item,
			recipes,
			character.ActiveItem,
		}

		itemDtos = append(itemDtos, itemDto)
	}

	return itemDtos
}

// Get the character's full inventory
func (model *Game) LoadInventory(character Character) []Inventory {
	inventory := NewInventory()
	inventoryList := inventory.GetInventoryForCharacter(int64(character.ID))

	return inventoryList
}

// Check if character has required items for the action
func (model *Game) CharacterHasItems(character Character, recipeItems []Recipe) string {
	inventory := NewInventory()
	if len(recipeItems) > 0 {
		for _, recipeItem := range recipeItems {
			inventoryItem := inventory.GetInventoryItemForCharacter(int64(recipeItem.ItemId), int64(character.ID))

			if inventoryItem.ID == 0 {
				return "You do not have enough resources to make this Item!"
			}

			if inventoryItem.Amount < recipeItem.Amount {
				return "Not enough resources to make this Item!"
			}
		}
	}

	return ""
}

// Remove Recipe Items from Inventory
func (model *Game) RemoveRecipeItemsFromInventory(recipeItems []Recipe, inventory Inventory, character Character) string {
	if len(recipeItems) > 0 {
		var inventoryItems []struct {
			Inventory
			Amount int64
		}

		for _, recipeItem := range recipeItems {
			inventoryItem := inventory.GetInventoryItemForCharacter(int64(recipeItem.ItemId), int64(character.ID))

			if inventoryItem.ID == 0 {
				return "You do not have enough resources to make this Item!"
			}

			if inventoryItem.Amount < recipeItem.Amount {
				return "You do not have enough resources to make this Item!"
			}

			inventoryItems = append(inventoryItems, struct {
				Inventory
				Amount int64
			}{
				inventoryItem,
				recipeItem.Amount,
			})
		}

		for _, inventoryItem := range inventoryItems {
			newAmount := inventoryItem.Inventory.Amount - inventoryItem.Amount
			if newAmount == 0 {
				inventoryItem.Delete()
			} else {
				inventoryItem.Inventory.Amount = newAmount
				inventoryItem.Update()
			}
		}
	}

	return ""
}

// EquipItem toggles one unit of an item between "carried" and "equipped".
//
// It used to move quantity out of the stack (Amount -= 1) and flip a boolean, which left
// empty rows behind and let a character wear two weapons. Now the row keeps the total and
// tracks how many of it are worn, and equipping one type unequips the others.
func (model *Game) EquipItem(characterId int64, itemId int64, inventory Inventory) string {
	inventoryItem := inventory.GetInventoryItemForCharacter(itemId, characterId)

	if inventoryItem.ID == 0 {
		return "You do not have the item in your inventory."
	}

	if !slices.Contains(EquippableItemTypes, inventoryItem.Item.Type) {
		return "That item is not equippable"
	}

	// Unequip one.
	if inventoryItem.EquippedAmount > 0 {
		model.db.Exec(
			`UPDATE inventories SET equipped_amount = 0, equipped = 0 WHERE id = ?`,
			inventoryItem.ID,
		)

		return ""
	}

	if inventoryItem.Amount <= 0 {
		return "You do not have the item in your inventory."
	}

	// One unit of each item may be worn, and different items may be worn together even of
	// the same type: several armour pieces at once are fine, two of the same named item are
	// not (clicking a worn item takes it off).
	model.db.Exec("UPDATE inventories SET equipped_amount = 1, equipped = 1 WHERE id = ?", inventoryItem.ID)

	return ""
}
