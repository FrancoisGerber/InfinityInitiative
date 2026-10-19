package tests

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func charactersFor(userID uint) []models.Character {
	m := models.NewCharacter()
	return m.GetCharactersForUser(int64(userID))
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
		Runtime:          0,
		RuntimeString:    "0s",
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
		Runtime:          0,
		RuntimeString:    "0s",
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

// doRequest issues a request and carries a refreshed session cookie forward, the way a
// browser would (pinning a character rewrites the session).
func doRequest(e http.Handler, method, path string, cookie **http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if *cookie != nil {
		req.AddCookie(*cookie)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			*cookie = c
		}
	}

	return rec
}

// Selecting a resource has to persist characters.active_item_id. It silently stopped
// doing so when updateAll() began omitting associations, because the handler only ever
// assigned character.ActiveItem.
func TestActiveItemSelectionPersists(t *testing.T) {
	user, character, gather, _ := gameFixture(t)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	rec := doRequest(e, http.MethodGet, "/game/action/"+strconv.FormatUint(uint64(gather.ID), 10), &cookie)
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

	// 5 coal pays for two crafts (2 + 2) and leaves one behind. Each craft is a start
	// click and a collect click: the first grants nothing.
	for i := 0; i < 2; i++ {
		rec := doRequest(e, http.MethodPost, craftPath, &cookie)
		if rec.Code != http.StatusOK {
			t.Fatalf("craft %d start returned %d: %q", i+1, rec.Code, rec.Body.String())
		}

		if got := inventoryItemFor(craft.ID, character.ID).Amount; got != int64(i) {
			t.Fatalf("the start click granted an item: rubies = %d, want %d", got, i)
		}

		rec = doRequest(e, http.MethodPost, "/game/action/collect", &cookie)
		if rec.Code != http.StatusOK {
			t.Fatalf("craft %d collect returned %d: %q", i+1, rec.Code, rec.Body.String())
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
	rec := doRequest(e, http.MethodPost, craftPath, &cookie)
	if !strings.Contains(strings.ToLower(rec.Body.String()), "resources") {
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
		rec := doRequest(e, http.MethodPost, "/game/equipitem/"+strconv.FormatUint(uint64(itemID), 10), &cookie)
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
	rec := doRequest(e, http.MethodPost, "/game/equipitem/"+strconv.FormatUint(uint64(axe.ID), 10), &cookie)
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

// Registration has to produce a playable account: a user with no character used to be
// written to game tables as character 0 (B8).
func TestRegistrationCreatesCharacter(t *testing.T) {
	e, _ := newTestServer()

	form := url.Values{
		"Email":         {"newbie@test.local"},
		"Username":      {"newbie"},
		"Password":      {"Passw0rd!123"},
		"CharacterName": {"Newbie"},
		"Class":         {"Mage"},
	}

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != 300 {
		t.Fatalf("register returned %d: %q", rec.Code, rec.Body.String())
	}

	userModel := models.NewUser()
	user := userModel.GetUserByEmail("newbie@test.local")
	if user.ID == 0 {
		t.Fatal("registration did not create the user")
	}

	characters := charactersFor(user.ID)
	if len(characters) != 1 {
		t.Fatalf("expected exactly one character, got %d", len(characters))
	}

	if characters[0].Class != "Mage" || characters[0].Name != "Newbie" {
		t.Fatalf("character not created as requested: %+v", characters[0])
	}
}

// An account can own several characters; the one pinned in the session is the one that
// gains items and XP (B7).
func TestActionsApplyToTheActiveCharacter(t *testing.T) {
	user, first, gather, _ := gameFixture(t)

	level := levelByName(1)
	second := models.InitCharacter(&models.Character{
		Name:          "Second",
		Class:         "Cleric",
		UserId:        user.ID,
		LevelId:       level.ID,
		CombatLevelId: level.ID,
		Health:        10,
		MaxHealth:     10,
	})
	if _, err := second.Create(); err != nil {
		t.Fatalf("second character: %v", err)
	}

	if got := len(charactersFor(user.ID)); got != 2 {
		t.Fatalf("expected two characters, got %d", got)
	}

	// Someone else's character id must not be switchable.
	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	rec := doRequest(e, http.MethodPost, "/game/character/"+strconv.FormatUint(uint64(second.ID), 10), &cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("switch returned %d", rec.Code)
	}
	if rec.Header().Get("HX-Refresh") != "true" {
		t.Fatal("switching did not ask the page to reload")
	}

	rec = doRequest(e, http.MethodPost, "/game/action/"+strconv.FormatUint(uint64(gather.ID), 10), &cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("action start returned %d: %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(e, http.MethodPost, "/game/action/collect", &cookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("action collect returned %d: %q", rec.Code, rec.Body.String())
	}

	inSecond := inventoryItemFor(gather.ID, second.ID)
	if inSecond.Amount != 1 {
		t.Fatalf("the active character did not get the item: amount = %d", inSecond.Amount)
	}

	inFirst := inventoryItemFor(gather.ID, first.ID)
	if inFirst.Amount != 0 {
		t.Fatalf("the inactive character was credited: amount = %d", inFirst.Amount)
	}
}

// The runtime must actually gate the payout: an action with a long runtime produces
// nothing until it has elapsed.
func TestActionRespectsRuntime(t *testing.T) {
	user, character, gather, _ := gameFixture(t)

	slow := models.InitItem(&models.Item{
		Name:             "Ancient Carrot",
		Type:             "Resource",
		LevelId:          gather.LevelId,
		ExperienceGained: 500,
		Runtime:          time.Hour,
		RuntimeString:    "1h",
		SkillId:          gather.SkillId,
	})
	if _, err := slow.Create(); err != nil {
		t.Fatalf("slow item: %v", err)
	}
	slow = itemByID(slow.ID)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)
	path := "/game/action/" + strconv.FormatUint(uint64(slow.ID), 10)

	if rec := doRequest(e, http.MethodPost, path, &cookie); rec.Code != http.StatusOK {
		t.Fatalf("start returned %d", rec.Code)
	}

	// Collecting before the runtime elapses grants nothing (the panel just keeps counting).
	if rec := doRequest(e, http.MethodPost, "/game/action/collect", &cookie); rec.Code != http.StatusOK {
		t.Fatalf("early collect returned %d", rec.Code)
	}

	if got := inventoryItemFor(slow.ID, character.ID).Amount; got != 0 {
		t.Fatalf("an action granted its output before its runtime elapsed: amount = %d", got)
	}

	// The pending action is recorded, so a later collect can finish it.
	reloaded := characterForUserID(user.ID)
	if reloaded.ActionItemId != slow.ID || reloaded.ActionReadyAt <= 0 {
		t.Fatalf("pending action not recorded: item=%d ready=%d", reloaded.ActionItemId, reloaded.ActionReadyAt)
	}
}
