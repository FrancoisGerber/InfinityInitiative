package models

import (
	"gorm.io/gorm"
)

type Role struct {
	Base
	gorm.Model
	ID   uint   `json:"ID" form:"ID" query:"ID"`
	Name string `json:"Name" form:"Name" query:"Name"`
}

func NewRole() Role {
	model := Role{}
	model.initDatabase()

	return model
}

func InitRole(model *Role) Role {
	model.initDatabase()

	return *model
}

func (model *Role) Migrate() error {
	migration := model.db.AutoMigrate(&model)

	admin := model.GetRoleByName("admin")
	if admin.ID == 0 {
		admin := NewRole()
		admin.Name = "admin"

		admin.Create()
	}

	user := model.GetRoleByName("user")
	if user.ID == 0 {
		user := NewRole()
		user.Name = "user"

		user.Create()
	}

	return migration
}

func (model *Role) GetAllRoles() (models []Role) {
	model.db.Find(&models)

	for _, m := range models {
		m = InitRole(&m)
	}

	return models
}

func (model *Role) GetRole(id int64) Role {
	var role Role
	model.db.First(&role, "id = ?", id)

	role = InitRole(&role)

	return role
}

func (model *Role) GetRoleByName(name string) Role {
	var role Role
	model.db.First(&role, "name = ?", name)

	role = InitRole(&role)

	return role
}

func (model *Role) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Role) Update() (int64, error) {
	result := model.db.Select("*").Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *Role) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
