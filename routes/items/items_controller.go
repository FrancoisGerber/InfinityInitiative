package items

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

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
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

	userId := c.Get("UserId").(uint)
	item := models.NewItem()
	items := item.GetAllItems()

	character := models.NewCharacter()

	data := models.Data{
		Page:       items,
		ActivePage: "items",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditItem(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

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

	return c.Render(http.StatusOK, "table-edit", data)
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
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

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

	return c.Render(http.StatusOK, "table-edit", data)
}

func ConfirmDeleteItem(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	item := models.NewItem()
	item = item.GetItem(Id)

	data := models.Data{
		Page: item,
	}

	return c.Render(http.StatusOK, "table-delete", data)
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
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

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

	return c.Render(http.StatusOK, "recipe-edit", data)
}

func AddRecipe(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})
	item := models.NewItem()
	items := item.GetAllItems()

	recipeDto := struct {
		models.Recipe
		Items []models.Item
	}{
		Recipe: models.Recipe{},
		Items:  items,
	}

	return c.Render(http.StatusOK, "recipe-edit-row", recipeDto)
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
