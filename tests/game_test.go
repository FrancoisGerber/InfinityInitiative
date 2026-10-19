package tests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"InfinityInitiative/data/models"
	"InfinityInitiative/utils"
)

// gameFixture creates a user, a character for it, one gatherable item and one item that
// costs 2 of the gatherable.
// Pointer-receiver getters need an addressable value; these keep the test bodies short.
func roleByName(name string) models.Role {
	m := models.NewRole()
	return m.GetRoleByName(name)
}

func levelByName(name int64) models.Level {
	m := models.NewLevel()
	return m.GetLevelByName(name)
}

func itemByID(id uint) models.Item {
	m := models.NewItem()
	return m.GetItem(int64(id))
}

func characterForUserID(userID uint) models.Character {
	m := models.NewCharacter()
	return m.GetCharacterForUser(int64(userID))
}

func inventoryItemFor(itemID, characterID uint) models.Inventory {
	m := models.NewInventory()
	return m.GetInventoryItemForCharacter(int64(itemID), int64(characterID))
}

func gameFixture(t *testing.T) (models.User, models.Character, models.Item, models.Item) {
	t.Helper()

	role := roleByName("User")
	if role.ID == 0 {
		t.Fatal("User role was not seeded")
	}

	handle := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))

	user := models.InitUser(&models.User{
		Username: handle,
		Email:    handle + "@test.local",
		Password: utils.HashPassword("Passw0rd!123"),
		RoleId:   role.ID,
	})
	if _, err := user.Create(); err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	level := levelByName(1)
	if level.ID == 0 {
		t.Fatal("level 1 was not seeded")
	}

	character := models.InitCharacter(&models.Character{
		Name:    "Gamer",
		UserId:  user.ID,
		LevelId: level.ID,
		// characters.combat_level_id is a NOT NULL foreign key with no default row.
		CombatLevelId: level.ID,
		Health:        10,
		MaxHealth:     10,
	})
	if _, err := character.Create(); err != nil {
		t.Fatalf("could not create character: %v", err)
	}

	gather := models.InitItem(&models.Item{
		Name:             "Coal",
		Type:             "Resource",
		LevelId:          int(level.ID),
		ExperienceGained: 100,
		Runtime:          time.Second,
		RuntimeString:    "1s",
		SkillId:          1,
	})
	if _, err := gather.Create(); err != nil {
		t.Fatalf("could not create gather item: %v", err)
	}
	gather = itemByID(gather.ID)

	craft := models.InitItem(&models.Item{
		Name:             "Ruby",
		Type:             "Resource",
		LevelId:          int(level.ID),
		ExperienceGained: 200,
		Runtime:          time.Second,
		RuntimeString:    "1s",
		SkillId:          5,
	})
	if _, err := craft.Create(); err != nil {
		t.Fatalf("could not create craft item: %v", err)
	}
	craft = itemByID(craft.ID)

	recipe := models.InitRecipe(&models.Recipe{ParentItemId: craft.ID, ItemId: gather.ID, Amount: 2})
	if _, err := recipe.Create(); err != nil {
		t.Fatalf("could not create recipe: %v", err)
	}

	return user, character, gather, craft
}

func doRequest(e http.Handler, method, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

// Selecting a resource has to persist characters.active_item_id. It silently stopped
// doing so when updateAll() began omitting associations, because the handler only ever
// assigned character.ActiveItem.
func TestActiveItemSelectionPersists(t *testing.T) {
	user, character, gather, _ := gameFixture(t)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	rec := doRequest(e, http.MethodGet, "/game/action/"+strconv.FormatUint(uint64(gather.ID), 10), cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("selecting the item returned %d: %q", rec.Code, rec.Body.String())
	}

	reloaded := characterForUserID(user.ID)
	if reloaded.ActiveItemId == nil {
		t.Fatalf("active_item_id was not persisted for character %d", character.ID)
	}
	if *reloaded.ActiveItemId != gather.ID {
		t.Fatalf("active_item_id = %d, want %d", *reloaded.ActiveItemId, gather.ID)
	}
}

// One click must consume the ingredients, grant the output and award XP — and a click
// that cannot be paid for must change nothing at all.
func TestRunActionIsAtomic(t *testing.T) {
	user, character, gather, craft := gameFixture(t)

	inventory := models.InitInventory(&models.Inventory{CharacterId: character.ID, ItemId: gather.ID, Amount: 5})
	if _, err := inventory.Create(); err != nil {
		t.Fatalf("could not seed inventory: %v", err)
	}

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)
	craftPath := "/game/action/" + strconv.FormatUint(uint64(craft.ID), 10)

	// 5 coal pays for two crafts (2 + 2) and leaves one behind.
	for i := 0; i < 2; i++ {
		rec := doRequest(e, http.MethodPost, craftPath, cookie)
		if rec.Code != http.StatusOK {
			t.Fatalf("craft %d returned %d: %q", i+1, rec.Code, rec.Body.String())
		}
	}

	remaining := inventoryItemFor(gather.ID, character.ID)
	if remaining.Amount != 1 {
		t.Fatalf("coal after two crafts = %d, want 1", remaining.Amount)
	}

	made := inventoryItemFor(craft.ID, character.ID)
	if made.Amount != 2 {
		t.Fatalf("rubies after two crafts = %d, want 2", made.Amount)
	}

	// The third click cannot be paid for: it must report the shortfall and leave the
	// bag exactly as it was.
	rec := doRequest(e, http.MethodPost, craftPath, cookie)
	if !strings.Contains(rec.Body.String(), "resources") {
		t.Fatalf("expected a shortfall message, got %q", rec.Body.String())
	}

	remaining = inventoryItemFor(gather.ID, character.ID)
	if remaining.Amount != 1 {
		t.Fatalf("a failed craft changed the ingredients: coal = %d, want 1", remaining.Amount)
	}

	made = inventoryItemFor(craft.ID, character.ID)
	if made.Amount != 2 {
		t.Fatalf("a failed craft still produced output: rubies = %d, want 2", made.Amount)
	}
}

// Equipping must not move quantity out of the stack, and only one item of a type may be
// worn at a time.
func TestEquipKeepsQuantityAndOneItemPerType(t *testing.T) {
	user, character, _, _ := gameFixture(t)

	level := levelByName(1)

	sword := models.InitItem(&models.Item{Name: "Bronze Sword", Type: "Weapon", LevelId: int(level.ID), SkillId: 4})
	if _, err := sword.Create(); err != nil {
		t.Fatalf("sword: %v", err)
	}
	sword = itemByID(sword.ID)

	axe := models.InitItem(&models.Item{Name: "Bronze Axe", Type: "Weapon", LevelId: int(level.ID), SkillId: 4})
	if _, err := axe.Create(); err != nil {
		t.Fatalf("axe: %v", err)
	}
	axe = itemByID(axe.ID)

	helm := models.InitItem(&models.Item{Name: "Bronze Helm", Type: "Armor", LevelId: int(level.ID), SkillId: 4})
	if _, err := helm.Create(); err != nil {
		t.Fatalf("helm: %v", err)
	}
	helm = itemByID(helm.ID)

	for _, itemID := range []uint{sword.ID, axe.ID, helm.ID} {
		row := models.InitInventory(&models.Inventory{CharacterId: character.ID, ItemId: itemID, Amount: 1})
		if _, err := row.Create(); err != nil {
			t.Fatalf("inventory row: %v", err)
		}
	}

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	for _, itemID := range []uint{sword.ID, helm.ID} {
		rec := doRequest(e, http.MethodPost, "/game/equipitem/"+strconv.FormatUint(uint64(itemID), 10), cookie)
		if rec.Code != http.StatusOK {
			t.Fatalf("equip %d returned %d", itemID, rec.Code)
		}
	}

	equippedSword := inventoryItemFor(sword.ID, character.ID)
	if equippedSword.Amount != 1 {
		t.Fatalf("equipping moved quantity: amount = %d, want 1", equippedSword.Amount)
	}
	if equippedSword.EquippedAmount != 1 {
		t.Fatalf("sword equipped_amount = %d, want 1", equippedSword.EquippedAmount)
	}

	// A second weapon replaces the first; the armour stays on.
	rec := doRequest(e, http.MethodPost, "/game/equipitem/"+strconv.FormatUint(uint64(axe.ID), 10), cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("equip axe returned %d", rec.Code)
	}

	swordAfter := inventoryItemFor(sword.ID, character.ID)
	if swordAfter.EquippedAmount != 0 {
		t.Fatalf("two weapons equipped at once: sword equipped_amount = %d", swordAfter.EquippedAmount)
	}

	helmAfter := inventoryItemFor(helm.ID, character.ID)
	if helmAfter.EquippedAmount != 1 {
		t.Fatalf("equipping a weapon unequipped the armour")
	}
}

func TestGameRepo(t *testing.T) {
	repo := models.NewGame()

	repo.Log("Hallo from testing")
}
