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
			"Items": loadFullItems(&skill, int64(userId)),
		},
		Character: models.GetCharacterForUser(int64(userId)),
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
	characterSkill := models.GetCharacterProgress(character.ID, uint(item.SkillId))

	nextLevelName := characterSkill.Level.Name + 1
	nextLevel := models.GetLevelByName(nextLevelName)

	nextCharacterLevelName := character.Level.Name + 1
	nextCharacterLevel := models.GetLevelByName(nextCharacterLevelName)

	skills := models.GetAllSkills()

	// Wait for action to run
	time.Sleep(item.Runtime)

	// TODO: Remove Recipe Items from Inventory
	// Check if the activeitem recipe items is in inventory
	// Remove recipe amounts from inventory

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
	character.ExperiencePoints = totalXP / int64(len(skills))

	if character.ExperiencePoints >= nextCharacterLevel.ExperienceNeeded {
		character.Level = nextCharacterLevel
	}
	_, err = character.Update()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	//TODO: Add Item to Character Inventory

	data := models.Data{
		Page: loadCharacterProgress(int64(userId), uint(item.SkillId)),
	}

	return c.Render(http.StatusOK, "action", data)
}

func GetActiveAction(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/game/*.html",
	})

	itemId := c.Param("itemId")
	item := models.GetItem(int64(utils.ParseInt(itemId)))
	userId := c.Get("UserId").(uint)
	character := models.GetCharacterForUser(int64(userId))

	//TODO: Check if item recipe resources is in inventory
	//TODO: Check if CharacterSkill level >= item level

	character.ActiveItem = item

	_, err := character.Update()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Update Character")
	}

	// data := models.Data{
	// 	Page: loadCharacterProgress(int64(userId), uint(item.SkillId)),
	// }

	// return c.Render(http.StatusOK, "action", data)

	c.Response().Header().Set("HX-Redirect", "/game/"+item.Skill.Name)
	c.Response().WriteHeader(http.StatusSeeOther)

	return nil
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

	characterSkillDto := struct {
		models.CharacterSkill
		NextLevel  models.Level
		ActiveItem models.Item
	}{
		characterSkill,
		nextLevel,
		activeItem,
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
