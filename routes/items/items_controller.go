package items

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/items/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/items", GetAll)
	g.GET("/items/:id", EditItem)
	g.POST("/items", SaveItem)
	g.GET("/items/create", CreateItem)
	g.GET("/items/delete/:id", ConfirmDeleteItem)
	g.DELETE("/items/:id", DeleteItem)
	g.GET("/items/recipe/:id", EditRecipe)
	g.POST("/items/recipe", SaveRecipe)
	g.GET("/items/recipe/add", AddRecipe)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)
	item := models.NewItem()
	items := item.GetAllItems()

	character := models.NewCharacter()

	data := models.Data{
		Page:       items,
		ActivePage: "items",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditItem(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	item := models.NewItem()
	level := models.NewLevel()
	skill := models.NewSkill()

	item = item.GetItem(Id)
	levels := level.GetAllLevels()
	skills := skill.GetAllSkills()

	data := models.Data{
		Page: item,
		DropdownData: map[string]any{
			"Levels":    levels,
			"Skills":    skills,
			"ItemTypes": models.ItemTypes,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveItem(c echo.Context) error {
	itemDto := new(models.Item)

	if err := c.Bind(itemDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error
	item := models.InitItem(itemDto)

	if item.RuntimeString != "" {
		runtimeDuration, err := time.ParseDuration(item.RuntimeString)

		if err != nil {
			return c.String(http.StatusBadRequest, "Cant Parse Runtime")
		}

		item.Runtime = runtimeDuration
	}

	// The form offers a <select>, but nothing validated the posted value: an unknown type
	// saved an item that could never be equipped or crafted, with no error to explain it.
	if !slices.Contains(models.ItemTypes, item.Type) {
		utils.ShowAlertError(c, "Unknown item type. Pick one of: "+strings.Join(models.ItemTypes, ", "))
		return nil
	}

	// A resource that belongs to a skill is an action. Without a runtime it resolves
	// instantly, so every click is free XP and a free copy of itself (B20).
	if item.Type == "Resource" && item.SkillId != 0 && item.Runtime <= 0 {
		utils.ShowAlertError(c, `A resource needs a runtime, e.g. "3s".`)
		return nil
	}

	if item.ID != 0 {
		_, err = item.Update()
	} else {
		_, err = item.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Item")
	}

	c.Response().Header().Set("HX-Redirect", "/items")
	c.Response().WriteHeader(300)

	return nil
}

func CreateItem(c echo.Context) error {

	level := models.NewLevel()
	skill := models.NewSkill()

	levels := level.GetAllLevels()
	skills := skill.GetAllSkills()

	data := models.Data{
		Page: models.Item{},
		DropdownData: map[string]any{
			"Levels":    levels,
			"Skills":    skills,
			"ItemTypes": models.ItemTypes,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func ConfirmDeleteItem(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	item := models.NewItem()
	item = item.GetItem(Id)

	data := models.Data{
		Page: item,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteItem(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	item := models.NewItem()
	item = item.GetItem(Id)

	_, err := item.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Item")
	}

	c.Response().Header().Set("HX-Redirect", "/items")
	c.Response().WriteHeader(300)

	return nil
}

func EditRecipe(c echo.Context) error {

	id := c.Param("id")
	ItemId := int64(utils.ParseInt(id))

	recipe := models.NewRecipe()
	item := models.NewItem()

	recipes := recipe.GetRecipesForParentItem(ItemId)
	items := item.GetAllItems()

	var recipesVM []struct {
		models.Recipe
		Items []models.Item
	}

	for _, recipe := range recipes {
		recipeVM := struct {
			models.Recipe
			Items []models.Item
		}{
			Recipe: recipe,
			Items:  items,
		}

		recipesVM = append(recipesVM, recipeVM)
	}

	data := models.Data{
		Page: recipesVM,
		Id:   ItemId,
	}

	return templates.RenderGroup(c, http.StatusOK, "recipe-edit", data, viewPatterns...)
}

func AddRecipe(c echo.Context) error {
	item := models.NewItem()
	items := item.GetAllItems()

	recipeDto := struct {
		models.Recipe
		Items []models.Item
	}{
		Recipe: models.Recipe{},
		Items:  items,
	}

	return templates.RenderGroup(c, http.StatusOK, "recipe-edit-row", recipeDto, viewPatterns...)
}

func SaveRecipe(c echo.Context) error {
	form, _ := c.FormParams()

	parentItemId := form["ParentItemID"][0]

	recipe := models.NewRecipe()
	_, err := recipe.DeleteRecipesOfParentItem(int64(utils.ParseInt(parentItemId)))

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Recipes for Parent Item"+err.Error())
	}

	for index, itemId := range form["ItemId"] {
		recipe := models.InitRecipe(&models.Recipe{
			ParentItemId: uint(utils.ParseInt(parentItemId)),
			ItemId:       uint(utils.ParseInt(itemId)),
			Amount:       int64(utils.ParseInt(form["Amount"][index])),
		})

		_, err = recipe.Create()

		if err != nil {
			return c.String(http.StatusInternalServerError, "Could Not Save Recipe")
		}
	}

	c.Response().Header().Set("HX-Redirect", "/items")
	c.Response().WriteHeader(300)

	return nil
}
