package models

type Game struct {
	Base
}

func NewGame(dbName string) Game {
	model := Game{}
	model.initDatabase(dbName)

	return model
}

// Get complete character skill progress
func (model *Game) LoadCharacterProgress(userId int64, skillId uint) struct {
	CharacterSkill
	NextLevel    Level
	ActiveItem   Item
	PercentageXP float64
} {

	character := NewCharacter(model.dbName)
	character = character.GetCharacterForUser(userId)
	characterSkill := GetCharacterProgress(character.ID, skillId)

	nextLevelName := characterSkill.Level.Name + 1
	nextLevel := GetLevelByName(nextLevelName)

	var activeItem Item
	if character.ActiveItem.SkillId == int(skillId) {
		activeItem = character.ActiveItem
	}

	newMax := nextLevel.ExperienceNeeded - characterSkill.Level.ExperienceNeeded
	newXp := characterSkill.ExperiencePoints - characterSkill.Level.ExperienceNeeded

	percentageXP := (float64(newXp) / float64(newMax) * 100)

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

	character := NewCharacter(model.dbName)
	character = character.GetCharacterForUser(userId)
	items := GetAllItemsForSkill(skill.ID)

	var itemDtos []struct {
		Item
		Recipes    []Recipe
		ActiveItem Item
	}

	for _, item := range items {
		recipes := GetRecipesForParentItem(int64(item.ID))

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
	character := NewCharacter(model.dbName)
	character = character.GetCharacterForUser(userId)
	inventory := GetInventoryForCharacter(int64(character.ID))

	return inventory
}

// Check if character has required items for the action
func (model *Game) CharacterHasItems(character Character, recipeItems []Recipe) string {
	if len(recipeItems) > 0 {
		for _, recipeItem := range recipeItems {
			inventoryItem := GetInventoryItemForCharacter(int64(recipeItem.ItemId), int64(character.ID))

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
