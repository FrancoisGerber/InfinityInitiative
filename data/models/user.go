package models

import (
	"InfinityInitiative/data"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uint   `json:"ID" form:"ID" query:"ID"`
	Username string `gorm:"unique" json:"Username" form:"Username" query:"Username"`
	Email    string `gorm:"unique" json:"Email" form:"Email" query:"Email"`
	Password string `json:"Password" form:"Password" query:"Password"`
	Image    string `json:"Image" form:"Image" query:"Image"`
	RoleId   uint   `json:"RoleId" form:"RoleId" query:"RoleId"`
	Role     Role   `json:"Role" form:"Role" query:"Role"`
}

func MigrateUser() {
	data.DB.AutoMigrate(&User{})
}

func GetAllUsers() (models []User) {
	data.DB.Find(&models)

	return models
}

func GetUser(id int64) (model User) {
	data.DB.First(&model, "id = ?", id)

	return model
}

func (model *User) Create() (int64, error) {
	result := data.DB.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *User) Update() (int64, error) {
	result := data.DB.Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *User) Delete() (int64, error) {
	result := data.DB.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
