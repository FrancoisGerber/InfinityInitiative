# Bugs — Infinity Initiative

Static audit of the implemented code (commit `bddb32b`) plus one live reproduction. Every entry was
read out of the source; nothing here is speculative. Line numbers are from that commit.

Severity: **Critical** = exploitable or destroys player data · **High** = wrong behaviour players
will hit · **Medium** = wrong under concurrency/edge cases · **Low** = hygiene.

## Critical

| ID | Where | Symptom | Cause | Fix |
|---|---|---|---|---|
| **B1** | `routes/router.go:21-37`, `utils/middleware.go:9` | **Any logged-in user is a full admin.** Verified live (see below): a `role = User` account got `200` on `/users`, `/roles`, `/items`, `/monsters`, `/characters` and `300` on `DELETE /items/:id`. | Every resource is mounted on `closedGroup`, which only runs `SessionAuth` (`userId != 0`). Nothing checks the role. The `RoleId == 1` check exists only in `sidebar.html` — cosmetic. | Add `utils.RequireRole(1)` middleware and mount admin resource groups behind it; keep `/game*` on the plain session group. |
| **B2** | `data/models/user.go:37-46` | Every boot creates `admin@admin.com` with password **`admin`** if it does not exist. | `User.Migrate()` seeds the admin, and `AutoMigration` runs for 10 models on startup. | Remove seeding from `Migrate`; make it a one-shot CLI/seed command reading credentials from env, and force a password change flag. |
| **B3** | `data/models/game.go:127-166` + `routes/game/game_controller.go:104-128` | Crafting is **non-atomic and non-idempotent**: ingredients are removed, then the handler sleeps, then XP and the output item are written. A crash, disconnect or second click mid-action loses ingredients and/or duplicates rewards. | No `db.Transaction`, no row lock, no request idempotency key. Two concurrent POSTs both pass `CharacterHasItems` and both pay out. | Wrap the whole action in one `db.Transaction`; take the ingredients `UPDATE ... WHERE amount >= ?` (affected-rows check) instead of read-then-write; key the action on `(character_id, request_id)` to reject replays. |
| **B4** | `routes/game/game_controller.go:110` | `time.Sleep(item.Runtime)` **inside the HTTP handler**, and runtimes are real: `items.runtime` stores nanoseconds, so Diamond Crystal = `60000000000` = **60 s**. | The action model resolves by blocking the request. | Never sleep in a request. Resolve the action immediately, and if pacing is wanted, gate on a stored `next_action_at` timestamp the client polls. |

## High

| ID | Where | Symptom | Cause | Fix |
|---|---|---|---|---|
| **B5** | `character.go:95-103`, `inventory.go:65-73`, `item.go`, `user.go`, `role.go`, `level.go`, `recipe.go` — every `Update()` | Updates can silently overwrite columns loaded moments earlier and can upsert related rows (`Character`, `Item`, `Level`). | `model.db.Select("*").Updates(&model)` — `model` is already a pointer, so this passes `**T`, and `Select("*")` forces **every** column (including zero values) plus preloaded associations. | `Updates(map[string]any{...})` for explicit columns, or `Select("amount").Updates(model)`. Never `Select("*")`. |
| **B6** | `utils/session.go:30-32, 65-67` | Session cookie has **no `HttpOnly`, `Secure`, `SameSite` or `MaxAge`**; no CSRF protection anywhere; the session is not rotated on login. | `sess.Options = &sessions.Options{Path: "/"}` is set fresh on every write. | One place builds cookie options (`HttpOnly: true, Secure: true, SameSite: Lax, MaxAge: 30d`); add `middleware.CSRFWithConfig`; regenerate the session id on login. |
| **B7** | `data/models/character.go:115-131` | Only the **first** character of an account is ever used; a second character is invisible and unplayable. Your DB already has two on `admin@admin.com` (ids 1 and 2). | `First(&character, "user_id = ?", userId)` = `ORDER BY id LIMIT 1`. Used by the game page, actions, sidebar, inventory and 5 admin controllers. | Add an active-character id to the session (or URL) plus `GetCharacterForUserAndId()` that puts the ownership check in the `WHERE`. |
| **B8** | `routes/auth/auth_controller.go:78-101` | A newly registered user has **zero characters**; `GetCharacterForUser` then returns `ID = 0`, and the game page happily writes `character_skills` and `inventories` rows for character 0. | Registration creates only the `users` row; characters are admin-created. | Create the first character during registration (or force a "create character" step before `/game`), and make `GetCharacterForUser` fail closed when none exists. |
| **B9** | `data/models/game.go:30-41` | XP bar math breaks at the level cap and for missing levels: `percentageXP = newXp / newMax * 100` where `newMax = nextLevel.ExperienceNeeded - level.ExperienceNeeded`. At level 99 there is no level 100 row → `nextLevel` is the zero value → divisor 0 or negative → `+Inf`/`NaN`/nonsense in the UI. | No guard on the last level and no "missing level" handling. | Clamp: if `nextLevel.ID == 0` treat as max level (100% bar, "MAX"); if `newMax <= 0` set 0. |
| **B10** | `game_controller.go:78,93` + `inventory.go:105-118` | "Your inventory is full!" fires when the bag is full **by row count**, even when the output would stack onto an existing row; and the check runs before the 1-60 s sleep, so it is stale by the time the item is added. `AddNewItemForCharacter` itself never checks the cap. | Capacity is `count(rows)` and the check is separated from the write. | Capacity = distinct stacks; enforce inside the transaction at insert time; allow stacking into existing rows regardless of cap. |
| **B11** | `data/models/inventory.go:95-105` | `GetInventoryItemForCharacter` returns an arbitrary row for `(item_id, character_id)` — after equip splits a stack into equipped + carried rows, "do I have this?" can read the 0-amount one. | `Find` with no `ORDER`/`LIMIT`, and the schema allows multiple rows per item. | One row per (character, item) with an `equipped` count, or `First(...).Order("equipped asc, id")` + explicit semantics. |
| **B12** | `data/models/game.go:226-250` | Equipping mutates the stack: `Amount -= 1` and a boolean flip on the same row. Consequences: zero-amount rows linger in the grid, you can equip two weapons, and unequipping a stack of 1 restores a phantom item. | `equipped` is a boolean on a *quantity* row. | Model equipment separately (`character_equipment(character_id, slot, item_id)`) or store `equipped_amount int`; validate slot type per `Item.Type`. |

## Medium

| ID | Where | Symptom | Cause | Fix |
|---|---|---|---|---|
| **B13** | `character_skill.go:86-110` | Opening a skill page **writes** a `character_skills` row. Two concurrent page loads can create duplicates (no unique index), and `First` then picks one arbitrarily. | Get-or-create in a read path with no constraint. | `UNIQUE(character_id, skill_id)` + `FirstOrCreate` inside a transaction. |
| **B14** | `game.go:169-198` | A single action can only advance **one** level, even if the XP crosses several thresholds; the remainder sits there until the next action. | A single `if`, not a loop. | `for xp >= next.ExperienceNeeded { level = next; next = next+1 }`. |
| **B15** | `game.go:185-191`, `character_skill.go:112-120` | Character level = `sum(skill XP) / len(allSkills)`. Adding an 8th skill retroactively changes every existing character's level; soft-deleted skill rows also count. | Hard-coded global divisor, raw SQL without `deleted_at IS NULL`. | Store `characters.experience_points` as its own value and add to it; or divide by skills the character actually has, with the soft-delete filter. |
| **B16** | 10 handlers, e.g. `game_controller.go:23,58,150` + `views/templates.go:20-38` | `c.Echo().Renderer = templates.ParseTemplates(...)` mutates the **shared** Echo instance per request: a data race, and under concurrency one request can render with another route's template set (wrong chrome or `no such template`). It also re-reads and re-parses ~60 files on every request (`template.Must` panics on a syntax error). | Renderer is global state; parse happens per request. | Cache one `*template.Template` per route group at startup (`sync.Once`), assign a per-request renderer via a context value or a small map keyed by layout. |
| **B17** | `utils/middleware.go:9-25` | A deleted user, or one demoted from Admin, keeps full access until the cookie is gone — the session is trusted forever. | `SessionAuth` only checks `userId != 0`; never loads the user or role. | Load the user (and role) per request, or store role+version in the session and bust it on change. |
| **B18** | `game_controller.go:104-146` | If the player navigates away mid-action the server still sleeps and still pays out — or, if the request is killed first, the ingredients are gone with no output (see B3). | Same root cause as B3/B4. | Transaction + no sleep; the client can then abandon safely. |
| **B19** | `utils/alert.go:19-21` + callers | Error paths render an OOB alert **and** return a plain-text body/status, so one response carries two payloads and htmx sometimes swaps the wrong fragment. | `ShowAlertError` renders immediately instead of returning data. | Handlers return `(fragment, status)`; one render per response. |
| **B20** | `views/items/table_edit.html:6` + `items_controller.go:85-93` | An item saved with an empty `RuntimeString` binds the hidden `Runtime=0` field → **instant** action. Any item whose `SkillId` is set appears as an action, so mis-entered equipment becomes free XP + a free copy of itself. | No validation that a skill action has a non-zero runtime / no separation of "action" vs "equipment". | Validate on save (`runtime > 0` required when `skill_id != 0`), and derive actionability from `type` rather than a stray `skill_id`. |
| **B21** | `app.go:30` | `middleware.CORS()` defaults to `AllowOrigins: *`. | Default config. | Explicit allow-list (or drop CORS entirely — this is a same-origin app). |
| **B22** | `utils/session.go:42-77` | UI state (`GameState` JSON) lives in the **cookie**; each write resets options; several more flags will hit the 4 KB cookie limit and silently drop the session. | Cookie store used as a scratchpad. | Server-side session store (or keep flags in `localStorage`). |

## Low / hygiene

| ID | Where | Issue |
|---|---|---|
| **B23** | `utils/session.go:21` | `userId.(uint)` unchecked type assertion — panics if the stored value ever deserialises as `int` (e.g. swapping the session store). |
| **B24** | `utils/middleware.go:19-23` | Handler errors are swallowed (`c.Error(err); return nil`), so failures log as 200 and the client sees a success. |
| **B25** | `data/models/data.go:15-25` | `Classes` lists 9 classes (Rogue, Druid, Paladin, Bard, Monk…) but only 4 icons exist (Cleric, Fighter, Mage, Ranger) and the seeded characters use Fighter. |
| **B26** | `data/models/data.go:3` | `MaxInventorySize` is a mutable package-level var (36); no per-character or per-upgrade override. |
| **B27** | `data/db_remote.go` | `InitRemote` is never called — the Turso/libsql dependency is dead weight in `go.mod`. |
| **B28** | `tests/` | Fixed-fixture tests against a gitignored DB: second run fails on `UNIQUE(users.email)`; `TestEnv` panics instead of `t.Fatal`; tests depend on the real `.env`; `go test`'s cache hides the failure. |
| **B29** | `Readme.md`, `views/assets/manifest.json` | README claims "Realtime Connection: Server-sent Events" — there is no SSE code. The manifest makes the app installable but there is **no service worker**, so "offline" is not true. |
| **B30** | repo root | `InfinityInitiative_backup.db`, `_Base.db`, `_LIVE.db`, `tests/InfinityInitiative.db` — gitignored but easy to confuse with the real DB; a stray `DBNAME` typo writes into one of them. |

## Verified reproduction — B1

A fresh `role = User` account (registered through the normal form, no admin involved) against the
running server:

```
/users 200   /roles 200   /items 200   /monsters 200   /characters 200
DELETE /items/999999 -> 300     (write endpoint reached, not 403)
```

Unauthenticated requests correctly return `307` to `/login`, so the session gate works — the missing
piece is purely the role check.

## Not bugs, but they will bite

- **`time.Sleep` + SQLite**: long blocking handlers plus a fresh connection per model construction
  (`data.Init()` inside every `NewX()`) is a recipe for `database is locked` under any concurrency.
  See `docs/improvements/06-data-layer-and-tests.md`.
- **Rewrite risk**: B3/B4/B5 are all "fix the action engine and the model write path" — doing them
  one at a time will thrash the same files. Sequence them as one piece of work.

## Fix status — branch `feat/game-systems`

**Fixed and verified (build + `go test -race` + a run against a copy of real data):**

| Bug | How it was closed |
|---|---|
| B1 | `routes/authz.go` `requireRole()` loads the user per request and returns 403; admin resources moved behind it. Regression tests: `tests/auth_test.go`. |
| B2 | Default admin seeding removed; first admin only via `SEED_ADMIN_EMAIL` + `SEED_ADMIN_PASSWORD` (12+ chars). |
| B3, B4, B18 | `data/models/action.go` `RunAction()`: one transaction, conditional ingredient decrement (`UPDATE ... WHERE amount >= ?`), output + XP inside it, **no sleep**. Measured: a 1 s-runtime action now resolves in ~5 ms. |
| B5 | `updateAll()` in `base.go` — explicit column list, associations omitted; all ten `Update()` methods use it. |
| B6 | Cookie is `HttpOnly`, `SameSite=Lax`, 30-day `MaxAge`, `Secure` unless `INSECURE_COOKIES=1`; `middleware.CSRF` guards every mutating request (header `X-CSRF-Token` from the `_csrf` cookie, injected by the `htmx:configRequest` hook in both shells); the session id rotates on login. A login POST without the token returns 400. |
| B9 | XP percentage clamped to 0–100 and guarded when there is no next level. |
| B10 | Capacity counts live stacks (`amount > 0`); stacking into an existing row is always allowed. |
| B11 | Inventory lookup is ordered (`equipped DESC, amount DESC, id ASC`); duplicate live stacks are merged by migration. |
| B12 | New `inventory.equipped_amount`; equipping no longer moves quantity, one item per type, zero-amount ghosts soft-deleted. |
| B13 | `UNIQUE(character_id, skill_id)` + `FirstOrCreate`. |
| B14 | Level-up loops until the XP no longer meets the next threshold. |
| B15 | Character XP is its own counter (+25 % of action XP) instead of `sum(skill XP)/len(allSkills)`. |
| B17 | The role middleware loads the user, so deleted/demoted accounts lose access immediately. |
| B21 | `middleware.CORS()` removed. |
| B24 | `SessionAuth` returns the handler's error instead of swallowing it. |
| B27 | `data/db_remote.go` deleted and `libsql-client-go` dropped from `go.mod`. |
| B28 | Tests get a throwaway DB per run (`TestMain`) and run from the repo root; suite passes repeatedly and under `-race`. |
| B7 | The session pins the active character; `GetCharacterForUserAndId` puts ownership in the `WHERE`; a sidebar panel lists the account's characters, switches them (`HX-Refresh`) and creates up to `MaxCharacters`. Every game lookup — progress, items, inventory, actions, equip, notes — now takes that character instead of `First(user_id)`. |
| B8 | Registration collects a character name + class and creates the character in the same request (user row rolled back if it fails); accounts that predate this get a default character on first page load instead of writing rows for character 0. |
| B16 | `templates.RenderGroup()` renders straight to the response; all ten controllers stopped assigning `c.Echo().Renderer` (zero mutations remain), and parsed sets are cached per pattern group. |
| B19 | The two remaining alert+`c.String` pairs return a single alert payload; `ShowAlert*` renders through the layout set. |
| B20 | Saving a `Resource` that belongs to a skill without a runtime is rejected, so no item can become a free instant action. |
| B22 | Sessions moved to a server-side filesystem store: the cookie carries a 160-byte id instead of the whole session (GameState included), so UI state can no longer hit the 4 KB cookie ceiling. |
| B23 | The stored user id is normalised for `uint/int/int64/float64` instead of a bare type assertion. |
| B25 | `Classes` trimmed to the four with portraits (Fighter, Mage, Cleric, Ranger). |
| B26 | `MaxInventorySize` is a `const`. |
| B29 | README no longer claims SSE and now states there is no service worker / offline mode. |
| B30 | A missing `DBNAME` is fatal instead of silently opening an unnamed database, the resolved path is logged at boot, and the stray `app-test.db` created during this audit was removed. The three pre-existing `*.db` backups in the repo stay (gitignored, real data). |

**Also fixed while in there:** duplicate live recipes are impossible (`UNIQUE(parent_item_id, item_id)`),
and every model shares one SQLite handle (WAL, `busy_timeout`, `MaxOpenConns(1)`) instead of opening a
pool per model construction — the data-layer half of the improvement plans.

**All audited bugs are closed.** Remaining work is the *design* backlog from `docs/game-systems/`
and `docs/improvements/` — there is no combat endpoint, and the 11 sidebar entries still have no
model, route or table.
