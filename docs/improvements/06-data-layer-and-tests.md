# 06 — Data layer and tests

**Fixes:** B5 (`Select("*").Updates(&model)`), B13 (duplicate `character_skills`), B27 (dead libsql),
B28 (test fixtures) · **Effort:** M

## One connection, not one per model

Every model constructor calls `data.Init()` → `gorm.Open(sqlite.Open(...))`, so a single game page
opens **5–10 separate pools** to the same file, per request. Combined with 60-second sleeps (B4) this
is how `database is locked` happens.

**Target:**

```go
// data/db.go
var once sync.Once
var conn *gorm.DB

func DB() *gorm.DB {          // returns the singleton
    once.Do(func() { conn = open(utils.DBName()) })
    return conn
}

func Tx(ctx context.Context, fn func(tx *gorm.DB) error) error  // the only way to write
```

- SQLite pragmas on open: `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`,
  `synchronous=NORMAL`.
- `SetMaxOpenConns(1)` for writes (SQLite serialises them anyway) or a small pool with WAL readers.
- Model constructors: `NewItem()` → `Item{Base: Base{db: data.DB()}}`; no I/O in a constructor.
- Delete `InitRemote`/`db_remote.go` and drop `libsql-client-go` from `go.mod` unless Turso is
  actually planned (B27).

## Writes

- **No `Select("*").Updates(&model)`.** It passes `**T`, writes zero values over good data, and with
  preloaded associations attached it upserts related rows. Replace with explicit column maps:
  `tx.Model(&inv).Updates(map[string]any{"amount": inv.Amount})`.
- **Every multi-row write goes through `Tx`** — the action engine (01), transfers (bank/market later),
  and admin save endpoints (a save with a recipe list is currently N independent writes).
- **Soft-delete filters in raw SQL**: `CalcCharacterTotalXP` sums deleted rows (B15);
  `GetInventoryForCharacterLen` correctly filters — make the rule uniform, or better, replace raw SQL
  with GORM queries that apply `deleted_at` automatically.

## Constraints (turn bugs into impossibilities)

| Table | Constraint |
|---|---|
| `character_skills` | `UNIQUE(character_id, skill_id)` — closes the duplicate-row race (B13) |
| `inventories` | `UNIQUE(character_id, item_id, location)` — one row per item, see plan 02 |
| `users` | `UNIQUE(email)` (already implied, make it explicit) |
| `recipes` | `UNIQUE(parent_item_id, item_id)`, `CHECK(amount > 0)` |
| `levels` | `UNIQUE(name)`, `CHECK(experience_needed >= 0)` |
| `characters` | `CHECK(health >= 0 AND health <= max_health)` once combat lands |

Write these as a real migration (gorm `AutoMigrate` will not add them to existing tables reliably —
use explicit `CREATE UNIQUE INDEX` statements in a numbered migration).

## Migrations and seeds

- Migrations: schema only, ordered, idempotent, run at boot; a `schema_migrations` table records what
  ran. Today `AutoMigration` runs for 10 models on every boot and `User.Migrate` writes data.
- Seeds: a separate `cmd/seed` with the admin user, roles, the 99 levels, 7 skills, and a starter item
  set. New content (monsters, dungeons) is admin-authored, not seeded.

## Tests

The current suite is one package against a gitignored fixture DB, and passes only on the first run.

- **Per-test database**: `t.TempDir()` + `DBNAME=...` set before `data.DB()` initialises, or an
  in-memory SQLite DSN (`file::memory:?cache=shared`). No test may touch a file another test uses.
- **Programmatic fixtures** (`tests/fixtures.go`): `NewTestUser`, `NewTestCharacter(itemSet)` — no
  fixed emails, no dependency on rows created by a previous run.
- **`t.Cleanup`** drops the schema; `TestMain` builds it once per package.
- **`make test` uses `-count=1`** so the Go cache cannot report a stale pass.
- **Regression tests for the bugs**: non-admin → 403 on `/users`; double action with one request id →
  single payout; level 99 → "MAX" not `NaN`; full bag + stackable output → accepted.
- **`go test -race ./...`** in CI — the renderer swap (B16) and the action race (B3) are only visible
  under `-race`.

## Makefile / CI

```
make ci    # go vet, go test -race -count=1 ./..., go build,
           # npm --prefix tailwindcss run prod && git diff --exit-code views/assets/css
```

The CSS check matters: `styles.css` is committed, and a template change without a rebuild ships stale
styles — a class of bug that is invisible in review.

## Acceptance criteria

- One DB handle per process (log opens at startup: exactly one).
- `go test -race -count=1 ./...` passes twice in a row, from a clean checkout, without touching
  `InfinityInitiative.db`.
- Every constraint in the table above exists in the migrated schema (`sqlite3 .indexes` shows them).
- Admin save of an item + recipes is atomic: a bad recipe row rolls back the item.
- `make ci` fails on a stale `styles.css`.
