package game

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/game/:skillName", GetGamePage)
	g.GET("/game/action/:itemId", GetActiveAction)
	g.POST("/game/action/:itemId", RunActiveAction)
	g.POST("/game/state/:name", SetGameState)
	g.POST("/game/notes", SetGameNotes)
	g.POST("/game/equipitem/:itemId", SetEquipedItem)
}

func GetGamePage(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	skillName := c.Param("skillName")

	game := models.NewGame()
	character := models.NewCharacter()
	skill := models.NewSkill()

	skill = skill.GetSkillByName(skillName)

	if skill.ID == 0 {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	userId := c.Get("UserId").(uint)

	data := models.Data{
		Page:       game.LoadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skillName,
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, int64(userId)),
			"Inventory":        game.LoadInventory(int64(userId)),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
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

	game := models.NewGame()

	// One atomic action: ingredients, XP and output in a single transaction, with no
	// sleep inside the request (see data/models/action.go).
	if _, err := game.RunAction(userId, uint(utils.ParseInt(itemId))); err != nil {
		utils.ShowAlertError(c, err.Error())
		return nil
	}

	item := models.NewItem()
	item = item.GetItem(int64(utils.ParseInt(itemId)))

	skill := models.NewSkill()
	skill = skill.GetSkill(int64(item.SkillId))

	character := models.NewCharacter()
	character = character.GetCharacterForUser(int64(userId))

	data := models.Data{
		Page:       game.LoadCharacterProgress(int64(userId), skill.ID),
		ActivePage: skill.Name,
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, int64(userId)),
			"Inventory":        game.LoadInventory(int64(userId)),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
	}

	c.Render(http.StatusOK, "oob-action-button", data)
	c.Render(http.StatusOK, "oob-resources", data)
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

	game := models.NewGame()
	character := models.NewCharacter()
	characterSkill := models.NewCharacterSkill()
	item := models.NewItem()
	recipe := models.NewRecipe()
	skill := models.NewSkill()

	// Load Data
	item = item.GetItem(int64(utils.ParseInt(itemId)))
	character = character.GetCharacterForUser(int64(userId))
	skill = skill.GetSkill(int64(item.SkillId))
	recipeItems := recipe.GetRecipesForParentItem(int64(item.ID))
	characterSkill = characterSkill.GetCharacterProgress(character.ID, skill.ID)

	// Check character skill level
	if characterSkill.Level.Name < item.Level.Name {
		utils.ShowAlertError(c, "You don't have the required level for this Item!")
		return c.String(http.StatusBadRequest, "You don't have the required level for this Item!!!!")
	}

	// Check if character has required items for the action
	if errorMessage := game.CharacterHasItems(character, recipeItems); errorMessage != "" {
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
			"Items":            game.LoadFullItems(&skill, int64(userId)),
			"Inventory":        game.LoadInventory(int64(userId)),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
	}

	// Load views
	c.Render(http.StatusOK, "oob-action-button", data)
	c.Render(http.StatusOK, "oob-resources", data)
	return c.Render(http.StatusOK, "oob-action", data)
}

func SetGameNotes(c echo.Context) error {
	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	character = character.GetCharacterForUser(int64(userId))

	character.Notes = c.FormValue("Notes")
	_, err := character.Update()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	return nil
}

func SetGameState(c echo.Context) error {
	gameState := utils.GetGameState(c)
	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	character = character.GetCharacterForUser(int64(userId))

	switch c.Param("name") {
	case "InventoryState":
		gameState.InventoryState = c.FormValue("InventoryState") == "on"
	case "NotesState":
		gameState.NotesState = c.FormValue("NotesState") == "on"
	case "ResourceListState":
		gameState.ResourceListState = c.FormValue("ResourceListState") == "on"
	case "InventoryTabState":
		gameState.InventoryTabState = c.FormValue("InventoryTabState") == "on"
		gameState.EquipmentTabState = false
		gameState.CharacterTabState = false
	case "EquipmentTabState":
		gameState.EquipmentTabState = c.FormValue("EquipmentTabState") == "on"
		gameState.InventoryTabState = false
		gameState.CharacterTabState = false
	case "CharacterTabState":
		gameState.CharacterTabState = c.FormValue("CharacterTabState") == "on"
		gameState.InventoryTabState = false
		gameState.EquipmentTabState = false
	}

	if err := utils.SetGameState(c, gameState); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	if c.Param("name") == "InventoryState" || c.Param("name") == "InventoryTabState" || c.Param("name") == "EquipmentTabState" || c.Param("name") == "CharacterTabState" {
		userId := c.Get("UserId").(uint)
		game := models.NewGame()

		// Get Viewmodel ready
		data := models.Data{
			DropdownData: map[string]any{
				"Inventory":        game.LoadInventory(int64(userId)),
				"GameState":        utils.GetGameState(c),
				"MaxInventorySize": models.MaxInventorySize,
			},
			Character: character.GetFullCharacterForUser(int64(userId)),
		}

		// Load views
		return c.Render(http.StatusOK, "oob-inventory", data)
	}

	return nil
}

func SetEquipedItem(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	itemId := c.Param("itemId")
	userId := c.Get("UserId").(uint)

	game := models.NewGame()
	character := models.NewCharacter()
	item := models.NewItem()
	skill := models.NewSkill()
	inventory := models.NewInventory()

	// Load Data
	item = item.GetItem(int64(utils.ParseInt(itemId)))
	character = character.GetCharacterForUser(int64(userId))

	// Equip the item
	msg := game.EquipItem(int64(character.ID), int64(item.ID), inventory)
	if msg != "" {
		utils.ShowAlertError(c, msg)
		return nil
	}

	// Get Viewmodel ready
	data := models.Data{
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, int64(userId)),
			"Inventory":        game.LoadInventory(int64(userId)),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.GetFullCharacterForUser(int64(userId)),
	}

	// Load views
	c.Render(http.StatusOK, "oob-equipped-items", data)
	return c.Render(http.StatusOK, "oob-inventory", data)
}
