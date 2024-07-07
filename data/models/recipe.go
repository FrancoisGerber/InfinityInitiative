package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type Recipe struct {
	gorm.Model
	ID           uint  `json:"ID" form:"ID" query:"ID"`
	ParentItemId uint  `json:"ParentItemId" form:"ParentItemId" query:"ParentItemId"`
	ParentItem   Item  `json:"ParentItem" form:"ParentItem" query:"ParentItem"`
	ItemId       uint  `json:"ItemId" form:"ItemId" query:"ItemId"`
	Item         Item  `json:"Item" form:"Item" query:"Item"`
	Amount       int64 `json:"Amount" form:"Amount" query:"Amount"`
}

func MigrateRecipes() {
	data.DB.AutoMigrate(&Recipe{})
}

func GetAllRecipes() (models []Recipe) {
	data.DB.Preload("Item").Preload("ParentItem").Find(&models)

	return models
}

func GetRecipe(id int64) (model Recipe) {
	data.DB.Preload("Item").Preload("ParentItem").First(&model, "id = ?", id)

	return model
}

func (model *Recipe) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Recipe) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func GetRecipesForParentItem(itemId int64) (models []Recipe) {
	data.DB.Preload("Item").Preload("ParentItem").Find(&models, "parent_item_id = ?", itemId)

	return models
}

func DeleteRecipesOfParentItem(itemId int64) (int64, error) {
	recipe := Recipe{}

	result := data.DB.Delete(&recipe, "parent_item_id = ?", itemId)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
