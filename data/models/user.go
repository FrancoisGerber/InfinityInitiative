package models

import (
	"InfinityInitiative/utils"

	"gorm.io/gorm"
)

type User struct {
	Base
	gorm.Model
	ID       uint   `json:"ID" form:"ID" query:"ID"`
	Username string `gorm:"unique" json:"Username" form:"Username" query:"Username"`
	Email    string `gorm:"unique" json:"Email" form:"Email" query:"Email"`
	Password string `json:"Password" form:"Password" query:"Password"`
	Image    string `json:"Image" form:"Image" query:"Image"`
	RoleId   uint   `json:"RoleId" form:"RoleId" query:"RoleId"`
	Role     Role   `json:"Role" form:"Role" query:"Role"`
}

func NewUser() User {
	model := User{}
	model.initDatabase()

	return model
}

func InitUser(model *User) User {
	model.initDatabase()

	return *model
}

func (model *User) Migrate() error {
	migration := model.db.AutoMigrate(&model)

	admin := model.GetUserByEmail("admin@admin.com")
	if admin.ID == 0 {
		admin := NewUser()
		admin.Username = "admin"
		admin.Email = "admin@admin.com"
		admin.Password = utils.HashPassword("admin")
		admin.RoleId = 1

		admin.Create()
	}

	return migration
}

func (model *User) GetAllUsers() (models []User) {
	model.db.Find(&models)

	for _, m := range models {
		m = InitUser(&m)
	}

	return models
}

func (model *User) GetUser(id int64) User {
	var user User
	model.db.First(&user, "id = ?", id)

	user = InitUser(&user)

	return user
}

func (model *User) GetUserByEmail(email string) User {
	var user User
	model.db.First(&user, "email = ?", email)

	user = InitUser(&user)

	return user
}

func (model *User) Create() (int64, error) {
	result := model.db.Create(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *User) Update() (int64, error) {
	result := model.db.Select("*").Updates(&model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *User) Delete() (int64, error) {
	result := model.db.Delete(&model, &model.ID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
