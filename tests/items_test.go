package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
)

// The admin form offers a <select> of item types, but nothing validated what was posted:
// an unknown type saved an item that could never be equipped or crafted, silently.
func TestSaveItemRejectsUnknownType(t *testing.T) {
	roleModel := models.NewRole()
	role := roleModel.GetRoleByName("Admin")
	if role.ID == 0 {
		t.Fatal("Admin role was not seeded")
	}

	handle := uniqueName("item-admin")

	user := models.InitUser(&models.User{
		Username: handle,
		Email:    handle + "@test.local",
		Password: utils.HashPassword("Passw0rd!123"),
		RoleId:   role.ID,
	})
	if _, err := user.Create(); err != nil {
		t.Fatalf("could not create admin: %v", err)
	}

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	itemModel := models.NewItem()
	before := len(itemModel.GetAllItems())

	level := levelByName(1)
	if level.ID == 0 {
		t.Fatal("level 1 was not seeded")
	}

	skillModel := models.NewSkill()
	skills := skillModel.GetAllSkills()
	if len(skills) == 0 {
		t.Fatal("no skills were seeded")
	}

	form := url.Values{
		"Name":    {"Typo Helm"},
		"Type":    {"Armour"}, // correct spelling is "Armor"
		"Image":   {"/assets/img/items/typo.png"},
		"LevelId": {strconv.FormatUint(uint64(level.ID), 10)},
		"SkillId": {strconv.FormatUint(uint64(skills[0].ID), 10)},
	}

	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if after := len(itemModel.GetAllItems()); after != before {
		t.Fatalf("unknown item type was saved: %d items before, %d after", before, after)
	}

	if !strings.Contains(rec.Body.String(), "Unknown item type") {
		t.Fatalf("rejection was not explained to the user; body: %.200s", rec.Body.String())
	}

	// A valid type still saves, so the check is a gate rather than a wall.
	form.Set("Name", "Proper Helm")
	form.Set("Type", "Armor")

	req = httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if after := len(itemModel.GetAllItems()); after != before+1 {
		t.Fatalf("a valid item was not saved: %d items before, %d after", before, after)
	}
}
