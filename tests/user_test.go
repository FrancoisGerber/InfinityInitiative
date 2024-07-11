package tests

import (
	"InfinityInitiative/data"
	"InfinityInitiative/data/models"
	"testing"
)

func TestUser(t *testing.T) {

	user := models.NewUser()
	data.AutoMigration(&user)

	u := models.InitUser(&models.User{
		Username: "Test",
		Email:    "test@test.com",
		Password: "test",
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

	u.Username = "Test2"
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
