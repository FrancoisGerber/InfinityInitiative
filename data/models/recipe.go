package models

import (
	"gorm.io/gorm"
)

type Recipe struct {
	Base
	gorm.Model
	ID           uint  `json:"ID" form:"ID" query:"ID"`
	ParentItemId uint  `json:"ParentItemId" form:"ParentItemId" query:"ParentItemId"`
	ParentItem   Item  `json:"ParentItem" form:"ParentItem" query:"ParentItem"`
	ItemId       uint  `json:"ItemId" form:"ItemId" query:"ItemId"`
	Item         Item  `json:"Item" form:"Item" query:"Item"`
	Amount       int64 `json:"Amount" form:"Amount" query:"Amount"`
}

func NewRecipe() Recipe {
	model := Recipe{}
	model.initDatabase()

	return model
}

func InitRecipe(model *Recipe) Recipe {
	model.initDatabase()

	return *model
}

func (model *Recipe) Migrate() error {
	if err := model.db.AutoMigrate(&model); err != nil {
		return err
	}

	// The recipe editor could save the same ingredient twice for one item, which made
	// crafting charge for it twice over. Keep the oldest row per (parent, ingredient),
	// then make it impossible to reintroduce.
	model.db.Exec(`
		DELETE FROM recipes
		WHERE deleted_at IS NULL
		  AND id NOT IN (
				SELECT MIN(id) FROM recipes
				WHERE deleted_at IS NULL
				GROUP BY parent_item_id, item_id
		  )
	`)

	return model.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_recipe_parent_item
		ON recipes(parent_item_id, item_id)
		WHERE deleted_at IS NULL
	`).Error
}

func (model *Recipe) GetAllRecipes() (models []Recipe) {
	model.db.Preload("Item").Preload("ParentItem").Find(&models)

	for _, m := range models {
		m = InitRecipe(&m)
	}

	return models
}

func (model *Recipe) GetRecipe(id int64) Recipe {
	var recipe Recipe
	model.db.Preload("Item").Preload("ParentItem").First(&recipe, "id = ?", id)

	recipe = InitRecipe(&recipe)

	return recipe
}

func (model *Recipe) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) Update() (int64, error) {
	result := updateAll(model.db, model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) Delete() (int64, error) {
	result := model.db.Delete(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) GetRecipesForParentItem(itemId int64) (models []Recipe) {
	model.db.Preload("Item").Preload("ParentItem").Find(&models, "parent_item_id = ?", itemId)

	for _, m := range models {
		m = InitRecipe(&m)
	}

	return models
}

func (model *Recipe) DeleteRecipesOfParentItem(itemId int64) (int64, error) {
	result := model.db.Delete(&model, "parent_item_id = ?", itemId)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
