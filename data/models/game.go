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

// Update CharacterSkill XP
func (model *Game) IncreaseCharacterSkill(characterSkill CharacterSkill, nextLevel Level, item Item) error {
	characterSkill.ExperiencePoints = characterSkill.ExperiencePoints + item.ExperienceGained

	if characterSkill.ExperiencePoints >= nextLevel.ExperienceNeeded {
		characterSkill.Level = nextLevel
	}
	_, err := characterSkill.Update()

	if err != nil {
		return err
	}

	return nil
}

// Update Character XP
func (model *Game) IncreaseCharacterXP(characterSkill CharacterSkill, character Character, nextCharacterLevel Level, skills []Skill) error {
	totalXP := characterSkill.CalcCharacterTotalXP(character.ID)
	character.ExperiencePoints = int64(float64(totalXP) / float64(len(skills)))

	if character.ExperiencePoints >= nextCharacterLevel.ExperienceNeeded {
		character.Level = nextCharacterLevel
	}
	_, err := character.Update()

	if err != nil {
		return err
	}

	return nil
}

// Add made item to inventory
func (model *Game) AddNewItemForCharacter(inventory Inventory, character Character, item Item) error {
	var err error
	if inventory.ID == 0 {
		inventoryItem := InitInventory(&Inventory{
			Character: character,
			Item:      item,
			Amount:    1,
			Equipped:  false,
		})

		_, err = inventoryItem.Create()
	} else {
		inventory.Amount += 1
		_, err = inventory.Update()
	}

	if err != nil {
		return err
	}

	return nil
}
