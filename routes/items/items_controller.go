package items

import (
	"InfinityInitiative/data/models"
	"InfinityInitiative/views"
	"net/http"
	"strconv"
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
}

func GetAll(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

	items := models.GetAllItems()

	data := models.Data{
		Page:       items,
		ActivePage: "items",
	}

	return c.Render(http.StatusOK, "index.html", data)
}

func EditItem(c echo.Context) error {
	c.Echo().Renderer = templates.ParseTemplates([]string{
		"views/layout/*.html",
		"views/items/*.html",
	})

	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	item := models.GetItem(Id)

	levels := models.GetAllLevels()

	data := models.Data{
		Page: item,
		DropdownData: map[string]any{
			"Levels":    levels,
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

	data := models.Data{
		Page: models.Item{},
		DropdownData: map[string]any{
			"Levels":    levels,
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
	Id, _ := strconv.ParseInt(id, 36, 64)

	item := models.GetItem(Id)

	data := models.Data{
		Page: item,
	}

	return c.Render(http.StatusOK, "table-delete", data)
}

func DeleteItem(c echo.Context) error {
	id := c.Param("id")
	Id, _ := strconv.ParseInt(id, 36, 64)

	item := models.GetItem(Id)

	_, err := item.Delete()

	if err != nil {
		return c.String(http.StatusInternalServerError, "Could Not Delete Item")
	}

	c.Response().Header().Set("HX-Redirect", "/items")
	c.Response().WriteHeader(300)

	return nil
}
