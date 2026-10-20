# 05 — Rendering and UX

**Status:** implemented on branch `improvements/05-rendering-and-ux` (2026-10-20). See
[Implementation](#implementation-2026-10-20) at the end for what changed, the two bugs found during the
live pass, and the verification run.

**Fixes:** B16 (per-request renderer swap + full re-parse), B19 (two payloads in one response), B22
(UI state in the cookie), B29 (claims the PWA does not deliver) · **Effort:** S–M

## Template pipeline

Every one of ~10 handlers does this, on the shared Echo instance:

```go
c.Echo().Renderer = templates.ParseTemplates([]string{"views/layout/*.html", "views/game/*.html"})
```

Three problems: it is a **data race** on global state (two requests can render with each other's
template sets), it re-reads and re-parses ~60 files **per request**, and `template.Must` turns a syntax
error into a 500 instead of a startup failure.

**Target:**

- Parse once at startup: one `*template.Template` per route group (`layout+game`, `layout+admin`,
  `auth`), stored in a small registry.
- Each handler picks a renderer from the registry and passes it via the context value Echo already
  supports (`c.Set("renderer", ...)`) — or, simpler, give each group its own `echo.Group` with the
  renderer assigned once.
- Keep a `make templates` / dev-mode re-parse behind a flag for hot reload (air already restarts on
  change, so this is optional).

Result: no race, no per-request file I/O, template errors at boot.

## Response contract

Right now an error path can render an OOB alert **and** return a plain-text body with a status
(`utils/alert.go` + most handlers), so htmx receives two payloads. Fix with one rule:

- a handler returns **one** fragment and one status code;
- errors are fragments too (an inline alert inside the panel that failed), never `c.String`;
- a small `render.Action(c, result)` helper does the OOB composition so handlers stop hand-assembling
  three `c.Render` calls.

## The clicker feel (small, high-value)

- **Indicators actually work.** `input.css` defines `.htmx-indicator` as an empty rule, so the loading
  bars in `action_button.html` never toggle. Define `display: none` + `.htmx-request .htmx-indicator {
  display: block }` — then it behaves as intended (and stops being permanently visible).
- **Optimistic numbers.** With plan 01 the action is instant; update the XP bar and item count in the
  same swap, and animate the delta client-side (a 150 ms CSS transition is enough).
- **A per-click log line** under the action panel: "+1 Oak Log · 25 XP", capped scrollback of ~20.
  This is the cheapest way to make a clicker feel alive.
- **Empty and error states**: "You need 5 more Coal" beats "Not enough resources". The data is already
  in the recipe check; surface the numbers.
- **Toasts** for success (`oob-alert-success` exists) instead of pushing content around.

## Layout and mobile

The current three-column layout (sidebar / action / resources+storage) is the right shape — it matches
the reference. Check it at 390 px: the daisyUI drawer already exists in `layout.html`, so the target is
sidebar → drawer, resources → tabs under the action, chat → hidden. One smoke pass with screenshots per
breakpoint, then fix what breaks.

Accessibility basics: every input has a label (currently the label is inside `.input` for the text
fields and above it for selects — make it consistent), buttons carry text or `aria-label`, and focus
rings survive the dark theme.

## PWA claims

`manifest.json` makes the app installable but there is no service worker, and the README claims
"Realtime Connection: Server-sent Events" while no SSE exists. Either:

- **honest minimum**: correct the README, keep the manifest (installable is nice), no offline claims;
- **or** add a tiny service worker that caches the shell + static assets only. **Do not** add an
  offline action queue — that is the background-timer model this game is defined against.

## Acceptance criteria

- `go test -race` with 50 concurrent page requests across two route groups: no race, no "no such
  template" errors.
- Template files are read once per process (verify with `strace -f -e openat` on a request, or a
  counter in the renderer).
- No handler calls both `c.Render` and `c.String` on the same request.
- Spinner appears on click and disappears on response.
- Layout usable at 390 px with no horizontal scroll.

## Implementation (2026-10-20)

**Template pipeline.** `templates.Preload()` is called at boot (`app.go`), and `templates.Groups()`
derives every route group from the views tree, so a template syntax error panics at startup instead of
500ing the first request. `views/templates.go` now has `Parse()` (returns the error) and `load()` (one
parse per group per process); `ParseCount()` exposes the counter the test asserts on. Handlers still
render through `RenderGroup`, never through `c.Echo().Renderer`.

**Response contract.** Every `c.String` error reply is gone (`grep -rn "c.String" --include=*.go` is
empty): auth, game and all eight admin controllers answer with a single alert fragment. The auth shell
gained the `#alerts` container it was missing, and every form whose response is an OOB swap carries
`hx-swap="none"` (`login`, `register`, all `table_edit.html`, `recipe_edit.html`), so an alert no longer
wipes the form it came from. `oob-alert-empty` was dead and is deleted.

**Two live bugs found and fixed along the way:**

- **The alert container deleted itself.** The OOB wrappers used `hx-swap="delete"` with the 5-second
  `/clearalert` timer, so after the first toast `#alerts` was gone and every later alert on that page
  silently vanished (htmx had no target). Clearing now swaps `innerHTML`; `ClearAlert` returns an empty
  200 and the container stays.
- **`HX-Trigger` was set after the first render.** `setActionLog` called `ShowAlertSuccess` and then set
  the header; `RenderGroup` had already written the status, so the header was dropped and level-ups got
  a toast but no log line. The header is now set before anything renders.

**Clicker feel.** `input.css` defines the real htmx indicator contract (hidden by default, `opacity: 1`
for `.htmx-request .htmx-indicator` / `.htmx-request.htmx-indicator`) and drops the stray
`.htmx-request{display:block !important}`; progress bars transition their fill over 150 ms. The action
panel carries a `#action-log` (`hx-preserve`, `aria-live="polite"`), hidden until the first result so
an empty panel shows no heading or placeholder. It is rendered with daisyUI's `list` /
`list-row` component: each resolved click sets an `HX-Trigger: actionResolved` event carrying the item
name, its icon, the XP and (on a level) the skill and new level, and the layout listener **prepends** a
row — icon, name, `+1` and `+200 XP` badges, plus a green skill badge and tinted row on a level-up.
Newest is on top, the list holds the last 20 and the oldest falls off the bottom, so the latest result
never scrolls out of view. Payload strings are escaped with `strconv.QuoteToASCII` because header
values decode as Latin-1. Success toasts now cover equips, character creation and level-ups.

**Errors that explain themselves.** `MissingIngredientError` (`data/models/action.go`) wraps
`ErrNotEnoughItems`, so both the click path (`checkAffordable`) and the spend path (`applyAction`,
affected-rows check) answer "You need 2 more Coal" instead of "not enough resources".
`CharacterHasItems` names the ingredient too.

**Layout, mobile, accessibility.** Verified at 390 px and 1280 px on a copied DB: no horizontal scroll
(`scrollWidth == clientWidth`), chat toggle hidden below `md`, everything stacks under the action card.
The first 390 px pass in Brave caught what the headless run missed: daisyUI menu items are
`flex: 0 0 auto`, so the two-row grid inside each resource row could not shrink and the page scrolled
sideways. Below 640 px the resource row is now a wrapping flex line (`input.css`). Character and chat
avatars use the local class portraits, so no image depends on the daisyUI stock CDN.
Sidebar toggle, chat, notifications and account menu have `aria-label`s; auth labels are bound with
`for`/`id`; resource rows and both action surfaces are keyboard reachable (`role="button"`,
`tabindex="0"`, `keyup[key=='Enter']`); focus-visible outlines survive the dark theme.

**Login redirect.** `PostLogin` used to `HX-Redirect` to `/characters`, the admin CRUD screen: every
non-admin landed on the 403 page. It redirects to `/`, which resolves to the first skill.

**Verification.** `go test -race ./tests/` passes with 50 concurrent requests across the game and items
groups (`TestConcurrentRequestsAcrossTemplateGroups`), one parse per group
(`TestTemplateSetsAreParsedOnce`), a single alert payload for a refused action
(`TestInsufficientIngredientsAnswerWithOneAlert`) and the ingredient name on the selection path. Live
pass on `DBNAME=/tmp/ii-05.db`: register → login → select Coal → click → log line, XP bar, level-up
toast, log capped at 20, indicator opacity 0→1, alert container survives clearing, a second alert still
lands, and both breakpoints screenshot clean.

