package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	ID   uint   `json:"ID" form:"ID" query:"ID"`
	Name string `json:"Name" form:"Name" query:"Name"`
}

func MigrateRoles() {
	data.DB.AutoMigrate(&Role{})
}

func GetAllRoles() (models []Role) {
	data.DB.Find(&models)

	return models
}

func GetRole(id int64) (model Role) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func (model *Role) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Role) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Role) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
