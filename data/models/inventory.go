package models

import (
	"gorm.io/gorm"
)

type Inventory struct {
	Base
	gorm.Model
	ID          uint      `json:"ID" form:"ID" query:"ID"`
	CharacterId uint      `json:"CharacterId" form:"CharacterId" query:"CharacterId"`
	Character   Character `json:"Character" form:"Character" query:"Character"`
	ItemId      uint      `json:"ItemId" form:"ItemId" query:"ItemId"`
	Item        Item      `json:"Item" form:"Item" query:"Item"`
	Equipped    bool      `json:"Equipped" form:"Equipped" query:"Equipped"`
	Amount      int64     `json:"Amount" form:"Amount" query:"Amount"`
}

func NewInventory() Inventory {
	model := Inventory{}
	model.initDatabase()

	return model
}

func InitInventory(model *Inventory) Inventory {
	model.initDatabase()

	return *model
}

func (model *Inventory) Migrate() error {
	return model.db.AutoMigrate(&model)
}

func (model *Inventory) GetAllInventorys() (models []Inventory) {
	model.db.Preload("Character").Preload("Item").Find(&models)

	for _, m := range models {
		m = InitInventory(&m)
	}

	return models
}

func (model *Inventory) GetInventory(id int64) Inventory {
	var inventory Inventory
	model.db.Preload("Character").Preload("Item").First(&inventory, "id = ?", id)

	inventory = InitInventory(&inventory)

	return inventory
}

func (model *Inventory) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) Update() (int64, error) {
	result := model.db.Select("*").Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) GetInventoryForCharacter(id int64) (models []Inventory) {
	model.db.Preload("Character").Preload("Item").Find(&models, "character_id = ?", id)

	for _, m := range models {
		m = InitInventory(&m)
	}

	return models
}

func (model *Inventory) GetInventoryItemForCharacter(itemId int64, characterId int64) Inventory {
	var inventory Inventory
	model.db.Preload("Character").Preload("Item").Find(&inventory, "item_id = ? and character_id = ?", itemId, characterId)

	inventory = InitInventory(&inventory)

	return inventory
}

func (model *Inventory) GetInventoryForCharacterLen(id int64) int64 {
	var Count int64

	model.db.Raw(`
		select count(id) as Count
		from inventories
		where character_id = ? and deleted_at is null
	`, id).Scan(&Count)

	return Count
}
