package recipes

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func SetupEndpoints(g *echo.Group) {
	g.GET("/recipes", GetAll)
	g.GET("/recipes/:id", EditRecipe)
	g.POST("/recipes", SaveRecipe)
	g.GET("/recipes/create", CreateRecipe)
	g.GET("/recipes/delete/:id", ConfirmDeleteRecipe)
	g.DELETE("/recipes/:id", DeleteRecipe)
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/recipes/*.html",
	})

	recipes := models.GetAllRecipes()

	data := models.Data{
		Page:       recipes,
		ActivePage: "recipes",
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditRecipe(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/recipes/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	recipe := models.GetRecipe(Id)
	items := models.GetAllItems()

	data := models.Data{
		Page: recipe,
		DropdownData: map[string]any{
			"Items": items,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func SaveRecipe(c echo.Context) error {
	recipe := new(models.Recipe)

	if err := c.Bind(recipe); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error

	if recipe.ID != 0 {
		_, err = recipe.Update()
	} else {
		_, err = recipe.Create()
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Save Recipe")
	}

	c.Response().Header().Set("HX-Redirect", "/recipes")
	c.Response().WriteHeader(300)

	return nil
}

func CreateRecipe(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/recipes/*.html",
	})

	items := models.GetAllItems()

	data := models.Data{
		Page: models.Recipe{},
		DropdownData: map[string]any{
			"Items": items,
		},
	}

	return c.Render(http.StatusOK, "table-edit", data)
}

func ConfirmDeleteRecipe(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/recipes/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	recipe := models.GetRecipe(Id)

	data := models.Data{
		Page: recipe,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteRecipe(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	recipe := models.GetRecipe(Id)

	_, err := recipe.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Recipe")
	}

	c.Response().Header().Set("HX-Redirect", "/recipes")
	c.Response().WriteHeader(300)

	return nil
}
