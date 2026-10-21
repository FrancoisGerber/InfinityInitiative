package tests

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
	"testing"
)

func TestUser(t *testing.T) {

	user := models.NewUser()
	data.AutoMigration(&user)

	roleModel := models.NewRole()
	userRole := roleModel.GetRoleByName("User")
	if userRole.ID == 0 {
		t.Fatal("User role was not seeded")
	}

	handle := uniqueName("test")

	u := models.InitUser(&models.User{
		Username: handle,
		Email:    handle + "@test.local",
		Password: utils.HashPassword("Passw0rd!123"),
		RoleId:   userRole.ID,
	})

	rows, err := u.Create()

	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	t.Log("CREATE Rows affected: ", rows)

	users := user.GetAllUsers()

	if len(users) == 0 {
		t.Log("Users is empty")
		t.Fail()
		return
	}

	t.Log("All Users: ", len(users))

	u.Username = uniqueName("test-updated")
	rows, err = u.Update()

	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	t.Log("UPDATE Rows affected: ", rows)

	rows, err = u.Delete()

	if err != nil {
		t.Log(err.Error())
		t.Fail()
		return
	}

	t.Log("DELETE Rows affected: ", rows)
}
