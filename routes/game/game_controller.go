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

	game := models.NewGame(utils.DBName)
	character := models.NewCharacter(utils.DBName)

	skill := models.GetSkillByName(skillName)

	if skill.ID == 0 {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       game.LoadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skillName,
		DropdownData: map[string]any{
			"Items":     game.LoadFullItems(&skill, int64(userId)),
			"Inventory": game.LoadInventory(int64(userId)),
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func RunActiveAction(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	itemId := c.Param("itemId")
	userId := c.Get("UserId").(uint)

	game := models.NewGame(utils.DBName)
	character := models.NewCharacter(utils.DBName)

	// Load Data
	item := models.GetItem(int64(utils.ParseInt(itemId)))
	character = character.GetCharacterForUser(int64(userId))
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
		Page:       game.LoadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skill.Name,
		DropdownData: map[string]any{
			"Items":     game.LoadFullItems(&skill, int64(userId)),
			"Inventory": game.LoadInventory(int64(userId)),
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
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

	itemId := c.Param("itemId")
	userId := c.Get("UserId").(uint)

	game := models.NewGame(utils.DBName)
	character := models.NewCharacter(utils.DBName)

	// Load Data
	item := models.GetItem(int64(utils.ParseInt(itemId)))
	character = character.GetCharacterForUser(int64(userId))
	skill := models.GetSkill(int64(item.SkillId))
	recipeItems := models.GetRecipesForParentItem(int64(item.ID))
	characterSkill := models.GetCharacterProgress(character.ID, skill.ID)

	// Check character skill level
	if characterSkill.Level.Name < item.Level.Name {
		utils.ShowAlertError(c, "You don't have the required level for this Item!")
		return c.String(http.StatusBadRequest, "You don't have the required level for this Item!!!!")
	}

	errorMessage := game.CharacterHasItems(character, recipeItems)

	if errorMessage != "" {
		utils.ShowAlertError(c, errorMessage)
		return c.String(http.StatusBadRequest, errorMessage)
	}

	// Set active action that may be performed
	character.ActiveItem = item

	_, err := character.Update()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	// Get viewmodel ready
	data := models.Data{
		Page: game.LoadCharacterProgress(int64(userId), uint(item.SkillId)),
		DropdownData: map[string]any{
			"Items":     game.LoadFullItems(&skill, int64(userId)),
			"Inventory": game.LoadInventory(int64(userId)),
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
	}

	// Load views
	c.Render(http.StatusOK, "oob-action-button", data)
	c.Render(http.StatusOK, "oob-recipes", data)
	return c.Render(http.StatusOK, "oob-action", data)
}
