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

	items := models.GetAllItems()

	character := models.NewCharacter(utils.DBName)

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

	item := models.GetItem(Id)

	levels := models.GetAllLevels()
	skills := models.GetAllSkills()

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
	item := new(models.Item)

	if err := c.Bind(item); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

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

	levels := models.GetAllLevels()
	skills := models.GetAllSkills()

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

	item := models.GetItem(Id)

	data := models.Data{
		Page: item,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteItem(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	item := models.GetItem(Id)

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

	recipes := models.GetRecipesForParentItem(ItemId)

	items := models.GetAllItems()

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
	items := models.GetAllItems()

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

	partenItemId := form["ParentItemID"][0]

	_, err := models.DeleteRecipesOfParentItem(int64(utils.ParseInt(partenItemId)))

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Recipes for Parent Item")
	}

	for index, itemId := range form["ItemId"] {
		recipe := models.Recipe{
			ParentItemId: uint(utils.ParseInt(partenItemId)),
			ItemId:       uint(utils.ParseInt(itemId)),
			Amount:       int64(utils.ParseInt(form["Amount"][index])),
		}

		_, err = recipe.Create()

		if err != nil {
			return c.String(http.StatusInternalServerError, "Could Not Save Recipe")
		}
	}

	c.Response().Header().Set("HX-Redirect", "/items")
	c.Response().WriteHeader(300)

	return nil
}
