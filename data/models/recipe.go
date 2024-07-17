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
	return model.db.AutoMigrate(&model)
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
	result := model.db.Select("*").Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

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
