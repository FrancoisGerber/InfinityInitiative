# 07 — Character switcher UI

**Status:** not implemented (deferred by the owner, 2026-10-19). **Effort:** S.

## The gap

Multi-character accounts work everywhere except in the UI. An account can own up to
`MaxCharacters` characters, `POST /game/character` creates one and `POST /game/character/:characterId`
switches the session-pinned active character — but **nothing in the interface links to either**,
so a player with two or more characters is stuck on whichever one the session pinned first.

Evidence:

- `routes/game/game_controller.go:27` mounts `GET /game/characters`, which renders
  `views/game/character-panel.html`.
- `grep -rn "game/characters\|character-panel" views/` finds **no reference** — the fragment is
  unreachable from any page.
- The data layer is complete and tested: `GetCharactersForUser`, `GetCharacterForUserAndId`
  (ownership in the `WHERE`), `CountForUser`, session helpers `GetSessionCharacter` /
  `SetSessionCharacter`, plus `TestActionsApplyToTheActiveCharacter`.

## Why it is not there

The panel existed in a first pass and was removed: it was injected into the sidebar, and the owner
rejected it there ("why is the side panel..."), so the markup went with it. The endpoints were left
behind, which is the state this plan closes.

## Constraint that shapes the fix

**Do not put it in the sidebar.** That placement was rejected once. A character belongs to the
account, not to the current page's navigation.

## Options

| Option | Shape | Cost | Notes |
|---|---|---|---|
| A. Character list page | `GET /characters` — a page listing each character (name, class, level), one "Play as" button per row, plus "Create character" | S | Reuses `character-panel.html` almost as-is; a page is where you would look for it |
| B. Header dropdown | A `<details>`/`<select>` in `layouts/layout.html` showing the active character's name, switching on change | S–M | Always visible, no page needed; needs an HTMX swap of the whole shell to avoid a stale sidebar |
| C. Navbar link to A | A single "Characters" link next to the account menu, page from A | S | Least surprise, ignores the rejected sidebar slot |

Recommended: **A + C**. The page is the feature; the link is `href`, no HTMX needed.

## Acceptance

1. A player with one character sees no switcher anywhere (no chrome for a non-choice).
2. A player with two or more can reach the list in one click from any game page.
3. "Play as" switches the session character and lands on the same skill page for the new
   character, whose inventory, skills, XP, storage and equipped items are visibly theirs.
4. "Create character" is hidden once the account owns `MaxCharacters`.
5. Switching never writes rows for character 0 (the failure mode B8 fixed).
6. The switch is a `POST` with CSRF, and the target id is checked against the account before use.

## Tests to add

- Switching sets the session character and the game page then renders that character's skill and
  inventory (extend `TestActionsApplyToTheActiveCharacter`).
- A character id owned by another account is refused (403/404, no switch).
- The create endpoint refuses a request past `MaxCharacters`.
