package models

import (
	"InfinityInitiative/utils"
	"os"

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
	model.seedAdmin()

	return migration
}

// seedAdmin creates the first administrator only when SEED_ADMIN_EMAIL and
// SEED_ADMIN_PASSWORD are set in the environment. There is no default account and no
// default password: a shipped "admin/admin" is an open door on any public deployment.
func (model *User) seedAdmin() {
	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")

	if email == "" || password == "" {
		return
	}

	if len(password) < 12 {
		model.Log("seed admin skipped: SEED_ADMIN_PASSWORD must be at least 12 characters")
		return
	}

	if model.GetUserByEmail(email).ID != 0 {
		return
	}

	admin := NewUser()
	admin.Username = "admin"
	admin.Email = email
	admin.Password = utils.HashPassword(password)
	admin.RoleId = 1

	if _, err := admin.Create(); err != nil {
		model.Log("seed admin failed: ", err.Error())
	}
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
	result := updateAll(model.db, model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (model *User) Delete() (int64, error) {
	result := model.db.Delete(model)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
