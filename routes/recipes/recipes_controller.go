package recipes

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	templates "InfinityInitiative/views"
	"net/http"

	"github.com/labstack/echo/v4"
)

// viewPatterns is the template set this controller renders into.
var viewPatterns = []string{"views/layout/*.html", "views/recipes/*.html"}

func SetupEndpoints(g *echo.Group) {
	g.GET("/recipes", GetAll)
	g.GET("/recipes/:id", EditRecipe)
	g.POST("/recipes", SaveRecipe)
	g.GET("/recipes/create", CreateRecipe)
	g.GET("/recipes/delete/:id", ConfirmDeleteRecipe)
	g.DELETE("/recipes/:id", DeleteRecipe)
}

func GetAll(c echo.Context) error {

	userId := c.Get("UserId").(uint)

	recipe := models.NewRecipe()
	character := models.NewCharacter()

	recipes := recipe.GetAllRecipes()

	data := models.Data{
		Page:       recipes,
		ActivePage: "recipes",
		Character:  character.GetFullCharacterForUser(int64(userId)),
	}

	return templates.RenderGroup(c, http.StatusOK, "index.html", data, viewPatterns...)
}

func EditRecipe(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	recipe := models.NewRecipe()
	item := models.NewItem()

	recipe = recipe.GetRecipe(Id)
	items := item.GetAllItems()

	data := models.Data{
		Page: recipe,
		DropdownData: map[string]any{
			"Items": items,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func SaveRecipe(c echo.Context) error {
	recipeDto := new(models.Recipe)

	if err := c.Bind(recipeDto); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var err error
	recipe := models.InitRecipe(recipeDto)

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

	item := models.NewItem()
	items := item.GetAllItems()

	data := models.Data{
		Page: models.Recipe{},
		DropdownData: map[string]any{
			"Items": items,
		},
	}

	return templates.RenderGroup(c, http.StatusOK, "table-edit", data, viewPatterns...)
}

func ConfirmDeleteRecipe(c echo.Context) error {

	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	recipe := models.NewRecipe()
	recipe = recipe.GetRecipe(Id)

	data := models.Data{
		Page: recipe,
	}

	return templates.RenderGroup(c, http.StatusOK, "table-delete", data, viewPatterns...)
}

func DeleteRecipe(c echo.Context) error {
	id := c.Param("id")
	Id := int64(utils.ParseInt(id))

	recipe := models.NewRecipe()
	recipe = recipe.GetRecipe(Id)

	_, err := recipe.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Recipe")
	}

	c.Response().Header().Set("HX-Redirect", "/recipes")
	c.Response().WriteHeader(300)

	return nil
}
