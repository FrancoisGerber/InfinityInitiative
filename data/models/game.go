package models

import "slices"

type Game struct {
	Base
}

func NewGame() Game {
	model := Game{}
	model.initDatabase()

	return model
}

// Get complete character skill progress
func (model *Game) LoadCharacterProgress(userId int64, skillId uint) struct {
	CharacterSkill
	NextLevel    Level
	ActiveItem   Item
	PercentageXP float64
} {

	character := NewCharacter()
	characterSkill := NewCharacterSkill()
	level := NewLevel()
	character = character.GetCharacterForUser(userId)
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

	characterSkillDto := struct {
		CharacterSkill
		NextLevel    Level
		ActiveItem   Item
		PercentageXP float64
	}{
		characterSkill,
		nextLevel,
		activeItem,
		percentageXP,
	}

	return characterSkillDto
}

// Get complete Items with recipes
func (model *Game) LoadFullItems(skill *Skill, userId int64) []struct {
	Item
	Recipes    []Recipe
	ActiveItem Item
} {

	item := NewItem()
	character := NewCharacter()
	character = character.GetCharacterForUser(userId)
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
func (model *Game) LoadInventory(userId int64) []Inventory {
	inventory := NewInventory()
	character := NewCharacter()
	character = character.GetCharacterForUser(userId)
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
			`UPDATE inventories SET equipped_amount = equipped_amount - 1,
			        equipped = (equipped_amount - 1) > 0
			 WHERE id = ?`, inventoryItem.ID,
		)

		return ""
	}

	if inventoryItem.Amount <= 0 {
		return "You do not have the item in your inventory."
	}

	// Equip one: only one item of each type may be worn at a time.
	var sameType []Inventory
	model.db.Preload("Item").
		Joins("JOIN items ON items.id = inventories.item_id").
		Where(`inventories.character_id = ? AND inventories.equipped_amount > 0
		       AND items.type = ? AND inventories.id <> ?`,
			characterId, inventoryItem.Item.Type, inventoryItem.ID).
		Find(&sameType)

	for _, other := range sameType {
		model.db.Exec("UPDATE inventories SET equipped = ?, equipped_amount = 0 WHERE id = ?", false, other.ID)
	}

	model.db.Exec("UPDATE inventories SET equipped_amount = equipped_amount + 1, equipped = ? WHERE id = ?", true, inventoryItem.ID)

	return ""
}
