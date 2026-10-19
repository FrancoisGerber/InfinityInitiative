# 05 — Rendering and UX

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
