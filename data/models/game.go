package models

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
