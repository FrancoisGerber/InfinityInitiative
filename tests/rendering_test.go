package tests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"InfinityInitiative/data/models"
	templates "InfinityInitiative/views"
)

// Every template group must parse once per process: the renderer used to re-read ~60
// files on every request, and the same files must not be parsed again later.
func TestTemplateSetsAreParsedOnce(t *testing.T) {
	if err := templates.Preload(); err != nil {
		t.Fatalf("preload failed: %v", err)
	}

	groups := templates.Groups()
	if len(groups) < 2 {
		t.Fatalf("expected at least two template groups, got %d", len(groups))
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

	handle := "concurrent-admin"
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

	if rec.Code != http.StatusOK {
		t.Fatalf("refused action returned %d: %q", rec.Code, body)
	}

	if !strings.HasPrefix(strings.TrimSpace(body), `<div id="alerts"`) {
		t.Fatalf("refused action did not return an alert fragment: %q", truncate(body, 200))
	}

	if strings.Count(body, `id="alerts"`) != 1 {
		t.Fatalf("response carried %d payloads, want 1: %q", strings.Count(body, `id="alerts"`), truncate(body, 300))
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

	if rec.Code != http.StatusOK {
		t.Fatalf("selecting the item returned %d: %q", rec.Code, body)
	}

	if want := "You need 2 " + gather.Name; !strings.Contains(body, want) {
		t.Fatalf("selection did not explain the shortfall (want %q): %q", want, truncate(body, 300))
	}
}
