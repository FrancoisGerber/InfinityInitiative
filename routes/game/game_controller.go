package game

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"InfinityInitiative/views"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/game/:skillName", GetGamePage)
	g.GET("/game/action/:itemId", GetActiveAction)
	g.POST("/game/action/:itemId", RunActiveAction)
}

func GetGamePage(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	skillName := c.Param("skillName")
	skill := models.GetSkillByName(skillName)

	if skill.ID == 0 {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       loadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skillName,
		DropdownData: map[string]any{
			"Items":     loadFullItems(&skill, int64(userId)),
			"Inventory": loadInventory(int64(userId)),
		},
		Character: models.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func RunActiveAction(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	// Load Data
	itemId := c.Param("itemId")
	item := models.GetItem(int64(utils.ParseInt(itemId)))
	userId := c.Get("UserId").(uint)

	character := models.GetCharacterForUser(int64(userId))

	inventoryCount := models.GetInventoryForCharacterLen(int64(character.ID))

	// Check if inventory is full
	if inventoryCount >= 20 {
		utils.ShowAlertError(c, "Your inventory if full!")
		return nil
	}

	characterSkill := models.GetCharacterProgress(character.ID, uint(item.SkillId))

	nextLevelName := characterSkill.Level.Name + 1
	nextLevel := models.GetLevelByName(nextLevelName)

	nextCharacterLevelName := character.Level.Name + 1
	nextCharacterLevel := models.GetLevelByName(nextCharacterLevelName)

	skills := models.GetAllSkills()
	skill := models.GetSkill(int64(item.SkillId))

	inventory := models.GetInventoryItemForCharacter(int64(item.ID), int64(character.ID))
	recipeItems := models.GetRecipesForParentItem(int64(item.ID))

	// Remove Recipe Items from Inventory
	if len(recipeItems) > 0 {
		var inventoryItems []struct {
			models.Inventory
			Amount int64
		}

		for _, recipeItem := range recipeItems {
			inventoryItem := models.GetInventoryItemForCharacter(int64(recipeItem.ItemId), int64(character.ID))

			if inventoryItem.ID == 0 {
				utils.ShowAlertError(c, "You do not have enough resources to make this Item!")
				return c.String(http.StatusBadRequest, "Not enough resources to make this Item!")
			}

			if inventoryItem.Amount < recipeItem.Amount {
				utils.ShowAlertError(c, "You do not have enough resources to make this Item!")
				return c.String(http.StatusBadRequest, "Not enough resources to make this Item!")
			}

			inventoryItems = append(inventoryItems, struct {
				models.Inventory
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

	// Wait for action to run
	time.Sleep(item.Runtime)

	// Update CharacterSkill XP
	characterSkill.ExperiencePoints = characterSkill.ExperiencePoints + item.ExperienceGained

	if characterSkill.ExperiencePoints >= nextLevel.ExperienceNeeded {
		characterSkill.Level = nextLevel
	}
	_, err := characterSkill.Update()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character Skill")
	}

	// Update Character XP
	totalXP := models.CalcCharacterTotalXP(character.ID)
	character.ExperiencePoints = int64(float64(totalXP) / float64(len(skills)))

	if character.ExperiencePoints >= nextCharacterLevel.ExperienceNeeded {
		character.Level = nextCharacterLevel
	}
	_, err = character.Update()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	// Add made item to inventory
	if inventory.ID == 0 {
		inventoryItem := models.Inventory{
			Character: character,
			Item:      item,
			Amount:    1,
			Equipped:  false,
		}

		_, err = inventoryItem.Create()
	} else {
		inventory.Amount += 1
		_, err = inventory.Update()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Inventory")
	}

	// Get Viewmodel ready
	data := models.Data{
		Page:       loadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skill.Name,
		DropdownData: map[string]any{
			"Items":     loadFullItems(&skill, int64(userId)),
			"Inventory": loadInventory(int64(userId)),
		},
		Character: models.GetFullCharacterForUser(int64(userId)),
	}

	// Load views
	c.Render(http.StatusOK, "oob-action", data)
	c.Render(http.StatusOK, "oob-side-bar", data)
	return c.Render(http.StatusOK, "oob-inventory", data)
}

func GetActiveAction(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	// Load Data
	itemId := c.Param("itemId")
	item := models.GetItem(int64(utils.ParseInt(itemId)))
	userId := c.Get("UserId").(uint)
	character := models.GetCharacterForUser(int64(userId))
	skill := models.GetSkill(int64(item.SkillId))
	recipeItems := models.GetRecipesForParentItem(int64(item.ID))
	characterSkill := models.GetCharacterProgress(character.ID, skill.ID)

	// Check character skill level
	if characterSkill.Level.Name < item.Level.Name {
		utils.ShowAlertError(c, "You don't have the required level for this Item!")
		return c.String(http.StatusBadRequest, "You don't have the required level for this Item!!!!")
	}

	// Check if character has required items for the action
	if len(recipeItems) > 0 {
		for _, recipeItem := range recipeItems {
			inventoryItem := models.GetInventoryItemForCharacter(int64(recipeItem.ItemId), int64(character.ID))

			if inventoryItem.ID == 0 {
				utils.ShowAlertError(c, "You do not have enough resources to make this Item!")
				return c.String(http.StatusBadRequest, "Not enough resources to make this Item!")
			}

			if inventoryItem.Amount < recipeItem.Amount {
				utils.ShowAlertError(c, "You do not have enough resources to make this Item!")
				return c.String(http.StatusBadRequest, "Not enough resources to make this Item!")
			}
		}
	}

	// Set active action that may be performed
	character.ActiveItem = item

	_, err := character.Update()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	//TODO: Start a background process to do this action automatically (IDLE mechanic)

	// Get viewmodel ready
	data := models.Data{
		Page: loadCharacterProgress(int64(userId), uint(item.SkillId)),
		DropdownData: map[string]any{
			"Items":     loadFullItems(&skill, int64(userId)),
			"Inventory": loadInventory(int64(userId)),
		},
		Character: models.GetFullCharacterForUser(int64(userId)),
	}

	// Load views
	c.Render(http.StatusOK, "oob-action-button", data)
	c.Render(http.StatusOK, "oob-recipes", data)
	return c.Render(http.StatusOK, "oob-action", data)
}

func loadCharacterProgress(userId int64, skillId uint) interface{} {
	character := models.GetCharacterForUser(userId)
	characterSkill := models.GetCharacterProgress(character.ID, skillId)

	nextLevelName := characterSkill.Level.Name + 1
	nextLevel := models.GetLevelByName(nextLevelName)

	var activeItem models.Item
	if character.ActiveItem.SkillId == int(skillId) {
		activeItem = character.ActiveItem
	}

	newMax := nextLevel.ExperienceNeeded - characterSkill.Level.ExperienceNeeded
	newXp := characterSkill.ExperiencePoints - characterSkill.Level.ExperienceNeeded

	percentageXP := (float64(newXp) / float64(newMax) * 100)

	characterSkillDto := struct {
		models.CharacterSkill
		NextLevel    models.Level
		ActiveItem   models.Item
		PercentageXP float64
	}{
		characterSkill,
		nextLevel,
		activeItem,
		percentageXP,
	}

	return characterSkillDto
}

func loadFullItems(skill *models.Skill, userId int64) interface{} {
	character := models.GetCharacterForUser(userId)
	items := models.GetAllItemsForSkill(skill.ID)

	var itemDtos []struct {
		models.Item
		Recipes    []models.Recipe
		ActiveItem models.Item
	}

	for _, item := range items {
		recipes := models.GetRecipesForParentItem(int64(item.ID))

		itemDto := struct {
			models.Item
			Recipes    []models.Recipe
			ActiveItem models.Item
		}{
			item,
			recipes,
			character.ActiveItem,
		}

		itemDtos = append(itemDtos, itemDto)
	}

	return itemDtos
}

func loadInventory(userId int64) []models.Inventory {
	character := models.GetCharacterForUser(userId)
	inventory := models.GetInventoryForCharacter(int64(character.ID))

	return inventory
}
