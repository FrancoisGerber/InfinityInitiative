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
	Equipped       bool      `json:"Equipped" form:"Equipped" query:"Equipped"`
	EquippedAmount int64     `json:"EquippedAmount" form:"EquippedAmount" query:"EquippedAmount"`
	Amount         int64     `json:"Amount" form:"Amount" query:"Amount"`
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
	if err := model.db.AutoMigrate(&model); err != nil {
		return err
	}

	// Older code could create several rows for the same item on one character (the
	// lookup used Find with no ordering, and equipping split stacks). Merge them so the
	// unique index below can exist and the grid shows real totals.
	model.mergeDuplicateStacks()

	// Ghost rows left by the old equip logic: zero quantity but still counted in the grid.
	model.db.Exec(`
		UPDATE inventories SET deleted_at = CURRENT_TIMESTAMP
		WHERE amount <= 0 AND deleted_at IS NULL
	`)

	// Partial unique index: soft-deleted rows must not collide.
	return model.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_character_item
		ON inventories(character_id, item_id)
		WHERE deleted_at IS NULL
	`).Error
}

// mergeDuplicateStacks collapses duplicate (character_id, item_id) rows into the oldest
// one: amounts are summed, an equipped flag on any duplicate is preserved.
func (model *Inventory) mergeDuplicateStacks() {
	model.db.Exec(`
		UPDATE inventories
		SET amount = (
				SELECT SUM(i2.amount) FROM inventories i2
				WHERE i2.character_id = inventories.character_id
				  AND i2.item_id = inventories.item_id
				  AND i2.deleted_at IS NULL
			),
			equipped = (
				SELECT MAX(i2.equipped) FROM inventories i2
				WHERE i2.character_id = inventories.character_id
				  AND i2.item_id = inventories.item_id
				  AND i2.deleted_at IS NULL
			)
		WHERE deleted_at IS NULL
		  AND id IN (
				SELECT MIN(id) FROM inventories
				WHERE deleted_at IS NULL
				GROUP BY character_id, item_id
		  )
	`)

	model.db.Exec(`
		DELETE FROM inventories
		WHERE deleted_at IS NULL
		  AND id NOT IN (
				SELECT MIN(id) FROM inventories
				WHERE deleted_at IS NULL
				GROUP BY character_id, item_id
		  )
	`)
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
	result := updateAll(model.db, model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Inventory) Delete() (int64, error) {
	result := model.db.Delete(model)

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

	// Deterministic pick: prefer the equipped row, then the largest stack. Previously
	// this used Find with no ordering and could return an empty duplicate row.
	model.db.Preload("Character").Preload("Item").
		Where("item_id = ? AND character_id = ?", itemId, characterId).
		Order("equipped DESC, amount DESC, id ASC").
		First(&inventory)

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
