package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"InfinityInitiative/data/models"
	templates "InfinityInitiative/views"
)

// Every template group must parse once per process: the renderer used to re-read ~60
// files on every request, and the same files must not be parsed again later.
func TestTemplateSetsAreParsedOnce(t *testing.T) {
	if err := templates.Preload(); err != nil {
		t.Fatalf("preload failed: %v", err)
	}

	groups, err := templates.Groups()
	if err != nil {
		t.Fatalf("listing template groups: %v", err)
	}

	if len(groups) < 2 {
		t.Fatalf("expected at least two template groups, got %d", len(groups))
	}

	// The layout group is fixed and must not be listed twice.
	seen := map[string]bool{}
	for _, group := range groups {
		key := strings.Join(group, "|")
		if seen[key] {
			t.Fatalf("template group %v is listed twice", group)
		}
		seen[key] = true
	}

	before := templates.ParseCount()

	for range 5 {
		for _, group := range groups {
			if templates.For(group...) == nil {
				t.Fatalf("group %v did not parse", group)
			}
		}
	}

	if after := templates.ParseCount(); after != before {
		t.Fatalf("repeated renders parsed %d template sets, want 0", after-before)
	}
}

// 50 concurrent requests across two route groups: the renderer is no longer mutated per
// request, so there must be no race and no "no such template" response.
func TestConcurrentRequestsAcrossTemplateGroups(t *testing.T) {
	role := roleByName("Admin")
	if role.ID == 0 {
		t.Fatal("Admin role was not seeded")
	}

	// Unique per run: a fixed handle made `-count=2` fail on users.email.
	handle := "concurrent-admin-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	admin := models.InitUser(&models.User{
		Username: handle,
		Email:    handle + "@test.local",
		Password: "not-used-here",
		RoleId:   role.ID,
	})
	if _, err := admin.Create(); err != nil {
		t.Fatalf("could not create admin: %v", err)
	}

	user, _, _, _ := gameFixture(t)

	e, store := newTestServer()
	adminCookie := signedSessionCookie(t, store, admin.ID)
	gameCookie := signedSessionCookie(t, store, user.ID)

	var wg sync.WaitGroup

	for i := range 50 {
		path := "/game/Mining"
		marker := "Mining"
		cookie := gameCookie
		if i%2 == 1 {
			path = "/items"
			marker = "Items"
			cookie = adminCookie
		}

		wg.Add(1)
		go func(path, marker string, cookie *http.Cookie) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(cookie)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s returned %d", path, rec.Code)
			}

			if !strings.Contains(rec.Body.String(), marker) {
				t.Errorf("%s rendered the wrong template set (missing %q): %q", path, marker, truncate(rec.Body.String(), 200))
			}
		}(path, marker, cookie)
	}

	wg.Wait()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n]
}

// A refused action must produce exactly one payload: an alert fragment the page can swap.
// It used to be able to render an alert and then return plain text in the same response.
func TestInsufficientIngredientsAnswerWithOneAlert(t *testing.T) {
	user, _, gather, craft := gameFixture(t)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	req := httptest.NewRequest(http.MethodPost, "/game/action/"+strconv.FormatUint(uint64(craft.ID), 10), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	body := rec.Body.String()

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("refused action returned %d, want 400: %q", rec.Code, body)
	}

	if rec.Header().Get("HX-Retarget") != "#alerts" {
		t.Fatalf("refused action is not retargeted at the toast container: %q", rec.Header().Get("HX-Retarget"))
	}

	if strings.Contains(body, `<div id="alerts"`) {
		t.Fatalf("refused action returned the OOB wrapper, not a single retargeted fragment: %q", truncate(body, 200))
	}

	if !strings.Contains(body, `role="alert"`) {
		t.Fatalf("refused action did not return an alert fragment: %q", truncate(body, 200))
	}

	if strings.Count(body, `role="alert"`) != 1 {
		t.Fatalf("response carried %d payloads, want 1: %q", strings.Count(body, `role="alert"`), truncate(body, 300))
	}

	// The retargeted fragment must carry its own clear timer; the static containers have
	// none, and without this the error toast stays forever.
	if !strings.Contains(body, `hx-get="/clearalert"`) || !strings.Contains(body, `hx-target="#alerts"`) {
		t.Fatalf("error alert does not re-arm the clear timer: %q", truncate(body, 300))
	}

	// The alert must name the ingredient and the shortfall, not say "not enough resources".
	want := "You need 2 more " + gather.Name
	if !strings.Contains(body, want) {
		t.Fatalf("alert did not explain the shortfall (want %q): %q", want, body)
	}
}

// Selecting a resource the character cannot afford explains the missing ingredients in
// the same way (the click path, not just the action engine).
func TestSelectingUnaffordableItemNamesTheIngredient(t *testing.T) {
	user, _, gather, craft := gameFixture(t)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	rec := doRequest(e, http.MethodGet, "/game/action/"+strconv.FormatUint(uint64(craft.ID), 10), &cookie)
	body := rec.Body.String()

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("selecting the item returned %d, want 400: %q", rec.Code, body)
	}

	if want := "You need 2 " + gather.Name; !strings.Contains(body, want) {
		t.Fatalf("selection did not explain the shortfall (want %q): %q", want, truncate(body, 300))
	}
}

// The action log is driven by the HX-Trigger header: assert it parses as JSON and names
// the item and XP, so the client contract is not left to the manual pass.
func TestActionTriggerCarriesTheLogPayload(t *testing.T) {
	user, _, gather, _ := gameFixture(t)

	e, store := newTestServer()
	cookie := signedSessionCookie(t, store, user.ID)

	path := "/game/action/" + strconv.FormatUint(uint64(gather.ID), 10)

	// The first click arms the action (runtime 0); the second collects it.
	doRequest(e, http.MethodPost, path, &cookie)
	rec := doRequest(e, http.MethodPost, path, &cookie)

	raw := rec.Header().Get("HX-Trigger")
	if raw == "" {
		t.Fatalf("collect did not send an HX-Trigger header: %q", truncate(rec.Body.String(), 200))
	}

	var payload struct {
		ActionResolved struct {
			Item  string `json:"item"`
			Image string `json:"image"`
			XP    int64  `json:"xp"`
		} `json:"actionResolved"`
	}

	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("HX-Trigger is not JSON (%v): %q", err, raw)
	}

	if payload.ActionResolved.Item != gather.Name {
		t.Fatalf("trigger item = %q, want %q", payload.ActionResolved.Item, gather.Name)
	}

	if payload.ActionResolved.XP != gather.ExperienceGained {
		t.Fatalf("trigger xp = %d, want %d", payload.ActionResolved.XP, gather.ExperienceGained)
	}

	if payload.ActionResolved.Image != gather.Image {
		t.Fatalf("trigger image = %q, want %q", payload.ActionResolved.Image, gather.Image)
	}
}

// A refused login keeps its HTTP status instead of answering 200 with a toast.
func TestFailedLoginKeepsItsStatus(t *testing.T) {
	role := roleByName("User")
	if role.ID == 0 {
		t.Fatal("User role was not seeded")
	}

	handle := "login-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	user := models.InitUser(&models.User{
		Username: handle,
		Email:    handle + "@test.local",
		Password: "hashed-elsewhere",
		RoleId:   role.ID,
	})
	if _, err := user.Create(); err != nil {
		t.Fatalf("could not create user: %v", err)
	}

	e, _ := newTestServer()

	form := url.Values{"Email": {handle + "@test.local"}, "Password": {"wrong-password"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("failed login returned %d, want 401: %q", rec.Code, truncate(rec.Body.String(), 200))
	}

	if !strings.Contains(rec.Body.String(), "Password is incorrect!") {
		t.Fatalf("failed login did not answer with the alert fragment: %q", truncate(rec.Body.String(), 200))
	}

	if !strings.Contains(rec.Body.String(), `hx-get="/clearalert"`) {
		t.Fatalf("login alert does not re-arm the clear timer: %q", truncate(rec.Body.String(), 200))
	}
}
