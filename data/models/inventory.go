package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type Inventory struct {
	gorm.Model
	ID          uint      `json:"ID" form:"ID" query:"ID"`
	CharacterId uint      `json:"CharacterId" form:"CharacterId" query:"CharacterId"`
	Character   Character `json:"Character" form:"Character" query:"Character"`
	ItemId      uint      `json:"ItemId" form:"ItemId" query:"ItemId"`
	Item        Item      `json:"Item" form:"Item" query:"Item"`
	Equipped    bool      `json:"Equipped" form:"Equipped" query:"Equipped"`
	Amount      int64     `json:"Amount" form:"Amount" query:"Amount"`
}

func MigrateInventorys() {
	data.DB.AutoMigrate(&Inventory{})
}

func GetAllInventorys() (models []Inventory) {
	data.DB.Preload("Character").Preload("Item").Find(&models)

	return models
}

func GetInventory(id int64) (model Inventory) {
	data.DB.Preload("Character").Preload("Item").First(&model, "id = ?", id)

	return model
}

func (model *Inventory) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func GetInventoryForCharacter(id int64) (models []Inventory) {
	data.DB.Preload("Character").Preload("Item").Find(&models, "character_id = ?", id)

	return models
}

func GetInventoryItemForCharacter(itemId int64, characterId int64) (model Inventory) {
	data.DB.Preload("Character").Preload("Item").Find(&model, "item_id = ? and character_id = ?", itemId, characterId)

	return model
}

func GetInventoryForCharacterLen(id int64) int64 {
	var Count int64

	data.DB.Raw(`
		select count(id) as Count
		from inventories
		where character_id = ?
	`, id).Scan(&Count)

	return Count
}
