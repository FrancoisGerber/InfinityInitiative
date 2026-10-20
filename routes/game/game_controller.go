package game

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/game/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/game/:skillName", GetGamePage)
	g.GET("/game/action/:itemId", GetActiveAction)
	g.POST("/game/action/:itemId", RunActiveAction)
	g.POST("/game/state/:name", SetGameState)
	g.POST("/game/notes", SetGameNotes)
	g.POST("/game/equipitem/:itemId", SetEquipedItem)
	g.POST("/game/collect", CollectActiveAction)
	g.GET("/game/characters", GetCharacterPanel)
	g.POST("/game/character", CreateCharacter)
	g.POST("/game/character/:characterId", SwitchCharacter)
}

// activeCharacter resolves the character this request is playing.
//
// The session pins the choice; without it the account's first character is used and pinned
// so the sidebar, actions and inventory all agree. Previously every lookup was
// First(user_id), which made a second character invisible and unplayable (B7).
func activeCharacter(c echo.Context, userId uint) models.Character {
	characterModel := models.NewCharacter()

	if characterId := utils.GetSessionCharacter(c); characterId != 0 {
		if character := characterModel.GetCharacterForUserAndId(int64(userId), int64(characterId)); character.ID != 0 {
			return character
		}
	}

	character := characterModel.GetCharacterForUser(int64(userId))
	if character.ID == 0 {
		// Accounts created before registration built a character would otherwise render
		// the page against character 0 and write rows for it. Give them the default.
		character = createDefaultCharacter(userId)
	}

	if character.ID != 0 {
		utils.SetSessionCharacter(c, character.ID)
	}

	return character
}

// createDefaultCharacter gives an account its first character (name = username).
func createDefaultCharacter(userId uint) models.Character {
	userModel := models.NewUser()
	user := userModel.GetUser(int64(userId))

	name := strings.TrimSpace(user.Username)
	if len(name) < 3 {
		name = "Adventurer"
	}
	if len(name) > 24 {
		name = name[:24]
	}

	level := models.NewLevel()
	firstLevel := level.GetLevelByName(1)

	character := models.InitCharacter(&models.Character{
		Name:          name,
		Class:         models.Classes[0],
		UserId:        userId,
		LevelId:       firstLevel.ID,
		CombatLevelId: firstLevel.ID,
		Health:        10,
		MaxHealth:     10,
		Attack:        1,
		Strength:      1,
		Defence:       1,
	})

	if _, err := character.Create(); err != nil {
		return models.Character{}
	}

	return character
}

// GetCharacterPanel renders the sidebar character list and the create-character form.
func GetCharacterPanel(c echo.Context) error {
	userId := c.Get("UserId").(uint)
	characterModel := models.NewCharacter()

	data := models.Data{
		DropdownData: map[string]any{
			"Characters": characterModel.GetCharactersForUser(int64(userId)),
			"ActiveId":   utils.GetSessionCharacter(c),
			"Classes":    models.Classes,
			"Max":        models.MaxCharacters,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "character-panel", data, viewPatterns...)
}

// SwitchCharacter pins another character of the same account to the session.
func SwitchCharacter(c echo.Context) error {
	userId := c.Get("UserId").(uint)

	characterModel := models.NewCharacter()
	character := characterModel.GetCharacterForUserAndId(int64(userId), int64(utils.ParseInt(c.Param("characterId"))))
	if character.ID == 0 {
		utils.ShowAlertError(c, "You do not have that character.")
		return nil
	}

	if err := utils.SetSessionCharacter(c, character.ID); err != nil {
		return err
	}

	// Reload: the page data underneath (skills, inventory, notes) belongs to the old one.
	c.Response().Header().Set("HX-Refresh", "true")

	return c.NoContent(http.StatusOK)
}

type newCharacterDto struct {
	Name  string
	Class string
}

// CreateCharacter adds another character to the account.
func CreateCharacter(c echo.Context) error {
	userId := c.Get("UserId").(uint)
	characterModel := models.NewCharacter()

	if characterModel.CountForUser(int64(userId)) >= models.MaxCharacters {
		utils.ShowAlertError(c, "This account already has the maximum number of characters.")
		return nil
	}

	dto := newCharacterDto{
		Name:  c.FormValue("Name"),
		Class: c.FormValue("Class"),
	}

	dto.Name = strings.TrimSpace(dto.Name)
	if len(dto.Name) < 3 || len(dto.Name) > 24 {
		utils.ShowAlertError(c, "Character name must be between 3 and 24 characters.")
		return nil
	}

	if !slices.Contains(models.Classes, dto.Class) {
		dto.Class = models.Classes[0]
	}

	level := models.NewLevel()
	firstLevel := level.GetLevelByName(1)

	character := models.InitCharacter(&models.Character{
		Name:          dto.Name,
		Class:         dto.Class,
		UserId:        userId,
		LevelId:       firstLevel.ID,
		CombatLevelId: firstLevel.ID,
		Health:        10,
		MaxHealth:     10,
		Attack:        1,
		Strength:      1,
		Defence:       1,
	})

	if _, err := character.Create(); err != nil {
		return err
	}

	if err := utils.SetSessionCharacter(c, character.ID); err != nil {
		return err
	}

	utils.ShowAlertSuccess(c, character.Name+" joined the Initiative.")

	return GetCharacterPanel(c)
}

func GetGamePage(c echo.Context) error {

	skillName := c.Param("skillName")

	game := models.NewGame()
	character := models.NewCharacter()
	skill := models.NewSkill()

	skill = skill.GetSkillByName(skillName)

	if skill.ID == 0 {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	userId := c.Get("UserId").(uint)
	character = activeCharacter(c, userId)

	data := models.Data{
		Page:       game.LoadCharacterProgress(character, skill.ID),
		ActivePage: skillName,
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, character),
			"Inventory":        game.LoadInventory(character),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.LoadFullCharacter(),
	}
	data = withPacing(data, character)

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func RunActiveAction(c echo.Context) error {
	itemID := uint(utils.ParseInt(c.Param("itemId")))
	userId := c.Get("UserId").(uint)

	game := models.NewGame()
	character := activeCharacter(c, userId)
	now := time.Now().UnixMilli()

	switch {
	case character.ActionItemId != 0 && now >= character.ActionReadyAt:
		// The runtime has elapsed; this click collects what the action produced.
		result, err := game.CollectAction(character.ID)
		if err != nil {
			utils.ShowAlertError(c, err.Error())
			return nil
		}

		setActionLog(c, result)

		character = activeCharacter(c, userId)

		// Keep the panel on the item that just finished.
		if itemID == 0 {
			itemID = result.Gained.ID
		}

		if itemID == 0 {
			itemID = character.ActionItemId
		}

		if itemID == 0 && character.ActiveItemId != nil {
			itemID = *character.ActiveItemId
		}
	case character.ActionItemId != 0:
		// Still running: re-render the countdown, grant nothing.
	default:
		if err := game.StartAction(character.ID, itemID); err != nil {
			if !errors.Is(err, models.ErrActionInProgress) {
				utils.ShowAlertError(c, err.Error())
				return nil
			}
		}

		character = activeCharacter(c, userId)
	}

	return renderActionFragments(c, userId, character, itemID)
}

// CollectActiveAction is the endpoint the action panel calls when the runtime elapses; it
// takes no item id, so an abandoned tab can still finish what it started.
func CollectActiveAction(c echo.Context) error {
	userId := c.Get("UserId").(uint)
	game := models.NewGame()
	character := activeCharacter(c, userId)

	result, err := game.CollectAction(character.ID)
	if err != nil {
		if !errors.Is(err, models.ErrActionInProgress) {
			utils.ShowAlertError(c, err.Error())
			return nil
		}

		return nil
	}

	setActionLog(c, result)

	character = activeCharacter(c, userId)

	// Render the panel around the item that was just collected, so it does not fall back to
	// "Select a resource" when the reset click never persisted a selection.
	itemID := result.Gained.ID
	if itemID == 0 {
		itemID = character.ActionItemId
	}
	if itemID == 0 && character.ActiveItemId != nil {
		itemID = *character.ActiveItemId
	}

	return renderActionFragments(c, userId, character, itemID)
}

// setActionLog sends the HX-Trigger event the page prepends to the action log. The event
// carries the item, its icon, the XP gained and (on a level-up) the skill and new level;
// the client renders the row with the daisyUI list component.
func setActionLog(c echo.Context, result models.ActionResult) {
	if result.Gained.Name == "" {
		return
	}

	skillName := ""
	if result.LeveledUp {
		skill := models.NewSkill()
		skill = skill.GetSkill(int64(result.SkillID))
		skillName = skill.Name
	}

	// Build the JSON by hand with QuoteToASCII: header values are decoded as Latin-1, so a
	// raw non-ASCII name would arrive mangled. JSON.parse restores the characters.
	payload := fmt.Sprintf(
		`{"actionResolved":{"item":%s,"image":%s,"xp":%d,"skill":%s,"level":%d,"leveledUp":%t}}`,
		strconv.QuoteToASCII(result.Gained.Name),
		strconv.QuoteToASCII(result.Gained.Image),
		result.Gained.ExperienceGained,
		strconv.QuoteToASCII(skillName),
		result.SkillLevel,
		result.LeveledUp,
	)

	// Headers must be set before anything is rendered: RenderGroup writes the status and
	// flushes, and a later Set() is ignored.
	c.Response().Header().Set("HX-Trigger", payload)

	if result.LeveledUp {
		utils.ShowAlertSuccess(c, fmt.Sprintf("%s is now level %d!", skillName, result.SkillLevel))
	}
}

// withPacing marks the data as "an action is running" so the action button renders as a
// countdown instead of a live button.
func withPacing(data models.Data, character models.Character) models.Data {
	now := time.Now().UnixMilli()
	if character.ActionItemId != 0 && character.ActionReadyAt > now {
		wait := character.ActionReadyAt - now
		data.Working = true
		data.WaitMs = wait
		data.WorkingSeconds = (wait + 999) / 1000
	}

	return data
}

// renderActionFragments sends the five out-of-band panels an action updates.
func renderActionFragments(c echo.Context, userId uint, character models.Character, itemID uint) error {
	game := models.NewGame()

	item := models.NewItem()
	item = item.GetItem(int64(itemID))

	skill := models.NewSkill()
	skill = skill.GetSkill(int64(item.SkillId))

	// The panel must show the item this click is about, even if the resource was never
	// "selected" (the navbar button posts straight to an action).
	if item.ID != 0 && (character.ActiveItemId == nil || *character.ActiveItemId != item.ID) {
		character.ActiveItem = item
		character.ActiveItemId = &item.ID
	}

	data := models.Data{
		Page:       game.LoadCharacterProgress(character, skill.ID),
		ActivePage: skill.Name,
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, character),
			"Inventory":        game.LoadInventory(character),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.LoadFullCharacter(),
	}
	data = withPacing(data, character)

	templates.RenderGroup(c, http.StatusOK, "oob-action-button", data, viewPatterns...)
	templates.RenderGroup(c, http.StatusOK, "oob-resources", data, viewPatterns...)
	templates.RenderGroup(c, http.StatusOK, "oob-action", data, viewPatterns...)
	templates.RenderGroup(c, http.StatusOK, "oob-side-bar", data, viewPatterns...)
	return templates.RenderGroup(c, http.StatusOK, "oob-inventory", data, viewPatterns...)
}

func GetActiveAction(c echo.Context) error {

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
	character = activeCharacter(c, userId)
	skill = skill.GetSkill(int64(item.SkillId))
	recipeItems := recipe.GetRecipesForParentItem(int64(item.ID))
	characterSkill = characterSkill.GetCharacterProgress(character.ID, skill.ID)

	// Check character skill level
	if characterSkill.Level.Name < item.Level.Name {
		// One response only: the alert fragment is the payload htmx swaps in.
		utils.ShowAlertError(c, "You don't have the required level for this Item!")
		return nil
	}

	// Check if character has required items for the action
	if errorMessage := game.CharacterHasItems(character, recipeItems); errorMessage != "" {
		utils.ShowAlertError(c, errorMessage)
		return nil
	}

	// Set active action that may be performed
	//
	// The foreign key must be set explicitly: updateAll() omits associations, so assigning
	// character.ActiveItem alone never reaches characters.active_item_id.
	character.ActiveItem = item
	character.ActiveItemId = &item.ID
	_, err := character.Update()
	if err != nil {
		utils.ShowAlertError(c, "Could not save the character. Try again.")
		return nil
	}

	// Get viewmodel ready
	data := models.Data{
		Page: game.LoadCharacterProgress(character, uint(item.SkillId)),
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, character),
			"Inventory":        game.LoadInventory(character),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.LoadFullCharacter(),
	}

	// Load views
	templates.RenderGroup(c, http.StatusOK, "oob-action-button", data, viewPatterns...)
	templates.RenderGroup(c, http.StatusOK, "oob-resources", data, viewPatterns...)
	return templates.RenderGroup(c, http.StatusOK, "oob-action", data, viewPatterns...)
}

func SetGameNotes(c echo.Context) error {
	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	character = activeCharacter(c, userId)

	character.Notes = c.FormValue("Notes")
	_, err := character.Update()
	if err != nil {
		utils.ShowAlertError(c, "Could not save the character. Try again.")
		return nil
	}

	return nil
}

func SetGameState(c echo.Context) error {
	gameState := utils.GetGameState(c)
	userId := c.Get("UserId").(uint)

	character := models.NewCharacter()
	character = activeCharacter(c, userId)

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
		utils.ShowAlertError(c, "Could not save the panel state. Reload the page.")
		return nil
	}

	if c.Param("name") == "InventoryState" || c.Param("name") == "InventoryTabState" || c.Param("name") == "EquipmentTabState" || c.Param("name") == "CharacterTabState" {
		userId := c.Get("UserId").(uint)
		character = activeCharacter(c, userId)
		game := models.NewGame()

		// Get Viewmodel ready
		data := models.Data{
			DropdownData: map[string]any{
				"Inventory":        game.LoadInventory(character),
				"GameState":        utils.GetGameState(c),
				"MaxInventorySize": models.MaxInventorySize,
			},
			Character: character.LoadFullCharacter(),
		}

		// Load views
		return templates.RenderGroup(c, http.StatusOK, "oob-inventory", data, viewPatterns...)
	}

	return nil
}

func SetEquipedItem(c echo.Context) error {

	itemId := c.Param("itemId")
	userId := c.Get("UserId").(uint)

	game := models.NewGame()
	character := models.NewCharacter()
	item := models.NewItem()
	skill := models.NewSkill()
	inventory := models.NewInventory()

	// Load Data
	item = item.GetItem(int64(utils.ParseInt(itemId)))
	character = activeCharacter(c, userId)

	// Equip the item
	msg := game.EquipItem(int64(character.ID), int64(item.ID), inventory)
	if msg != "" {
		utils.ShowAlertError(c, msg)
		return nil
	}

	utils.ShowAlertSuccess(c, "Equipped "+item.Name+".")

	// Get Viewmodel ready
	data := models.Data{
		DropdownData: map[string]any{
			"Items":            game.LoadFullItems(&skill, character),
			"Inventory":        game.LoadInventory(character),
			"GameState":        utils.GetGameState(c),
			"MaxInventorySize": models.MaxInventorySize,
		},
		Character: character.LoadFullCharacter(),
	}

	// Load views
	return templates.RenderGroup(c, http.StatusOK, "oob-inventory", data, viewPatterns...)
}
