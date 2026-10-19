# AGENTS.md

Context for any agent working in this repo. Read this before changing code.

## What this is

**Infinity Initiative** — a browser, text-based fantasy idle MMORPG.
Go + Echo + HTMX + SQLite/GORM, server-rendered `html/template` views, Tailwind CSS 4 + daisyUI 5 for styling, compiled into a committed bundle.

## Commands

| Task | Command |
|---|---|
| Dev server (hot reload) | `make dev` |
| Build / run | `make build`, `make run` |
| Tests | `make test` (see the test-DB caveat below) |
| CSS (watch) | `npm --prefix tailwindcss install && npm --prefix tailwindcss run dev` |
| CSS (minified, commit the output) | `npm --prefix tailwindcss run prod` |

The server listens on `:8069` (hardcoded in `app.go`).

## Architecture map

- `app.go` — bootstrap, `AutoMigration` for every model, Echo start.
- `routes/router.go` — mounts an open group (auth) and a closed group (everything else); one package per resource under `routes/*/`.
- `data/models/` — GORM models *and* their queries (`base.go` / `data.go` are the helpers). No separate repository layer.
- `views/` — templates. `layout/layout.html` + `layout/sidebar.html` are the shell; `*_table_edit.html` and `*_table_row.html` are HTMX fragments loaded into modals — opened directly on their own URL they render unstyled, so test them through the UI that fetches them.
- `utils/` — session, game state, formatters.
- `tailwindcss/input.css` — the whole CSS entrypoint (config is CSS-first; there is no `tailwind.config.js`).

## Implemented vs placeholder

**Implemented:** skills with XP-per-action, recipes/crafting, items with equip bonuses, inventory/storage, characters, users/roles, notes, monsters as admin CRUD.

**Placeholder** — sidebar entries that are bare `<a>` with no `href`, and have no model, route, or table: Quests, Raids, Dungeons, Market, Shop, Initiatives, House, Bank, Battle PvE, Initiative War PvP. Do not assume behaviour for these; they are unbuilt plans.

Per-entry implementation plans — data model, routes, order of work, and the decisions still open — live in `docs/game-systems/` (index: `README.md`). Read the matching plan before writing code for any of those entries.

Combat *stats* exist on `characters` (`Health`, `MaxHealth`, `Attack`, `Strength`, `Defence`, `ArmourRating`) but there is **no combat endpoint** — `routes/game` is only `/game/:skillName`, `/game/action/:itemId`, `/game/state/:name`, `/game/notes`, `/game/equipitem/:itemId`.

## Glossary

### Initiative

**An Initiative is this game's guild** — a player organisation. It is *not* a turn-order stat and *not* a quest chain.

- The sidebar placement is the evidence: `Initiatives` sits under **Locations** (with Market / Shop / House), so it is a place you visit; `Initiative War PvP` sits under **Combat**, so a war is the mode you fight in.
- Intended loop: members contribute gathered/crafted resources to level their Initiative → the shared hall (the counterpart to the personal `House`) → Initiatives fight each other in **Initiative Wars**, resolved as ordered rounds using the existing Attack / Strength / Defence / Armour Rating stats, so a war can resolve offline and produce a battle log. "Initiative" in the war's name is the turn order those rounds follow.
- The project name reads the same way: *Infinity Initiative* is the player's own outfit, forever.
- **Status: concept agreed, planned** — full plan in `docs/game-systems/07-initiatives.md`. Nothing is implemented: no model, route, table, or template.
- Rejected reading: "initiatives" as server-wide community goals, or as quests (`Quests` already exists as its own entry).

## Gotchas learned the hard way

- **Test DB is not idempotent.** `tests/InfinityInitiative.db` is a gitignored fixture; `TestUser` inserts a fixed `test@test.com`, so every run after the first fails with `UNIQUE constraint failed: users.email`. Remove the file, or point `DBNAME` at a temp path. `make test` can also look green from Go's test cache — use `-count=1` when you need a real run.
- **Never let Tailwind scan its own output.** `input.css` carries `@source not "../views/assets/css"` (and `.../assets/js`); without it, every emitted class becomes a permanent candidate and dead classes never disappear from the bundle.
- **Tailwind 4 browser floor:** Chrome/Safari 16.4+/Firefox 128+.
- **Experiments get a throwaway DB:** `DBNAME=/tmp/ii-scratch.db ./InfinityInitiative`. The repo's `InfinityInitiative.db` holds real data; `.env` holds `DBNAME` + a Turso token and must never be committed.
- `tailwindcss/node_modules` is gitignored; `views/assets/css/styles.css` is committed and must be regenerated with `npm run prod` whenever templates or `input.css` change.
