# 04 — Security and admin

**Fixes:** B1 (anyone is admin — verified live), B2 (seeded `admin`/`admin`), B6 (cookie/CSRF), B17
(stale sessions), B24 (errors swallowed) · **Effort:** M

## B1 first, because it is live

`SessionAuth` proves *someone* is logged in; nothing proves *who* they are. All nine resource packages
are mounted on that group, so a `role = User` account can read and write items, monsters, users,
roles, levels and recipes. Verified with a freshly registered non-admin account: `200` on every admin
page and `300` on `DELETE /items/:id`.

**Fix:**

```go
// utils/middleware.go
func RequireRole(roleID uint) echo.MiddlewareFunc   // loads the user, compares RoleId, 403 page otherwise
```

- Mount `users, roles, levels, skills, items, recipes, characters, monsters` on
  `closedGroup.Group("", RequireRole(1))`.
- Keep `game` on the plain session group.
- Load the **user row** per request (it is one indexed lookup) instead of trusting the cookie for
  identity — that also fixes B17 (deleted or demoted users keep access until logout).
- Render a real 403 template, not a redirect; htmx requests get a fragment with an error toast.

## Then remove the back door

`User.Migrate()` seeds `admin@admin.com` / `admin` on every boot. Replace with:

- a `cmd/seed` (or `make seed-admin`) that reads `ADMIN_EMAIL`/`ADMIN_PASSWORD` from env, refuses
  passwords shorter than 12 chars, and prints a one-time warning;
- a `must_change_password` flag on `users` so the first login forces a rotation;
- no seeding inside migrations at all (migrations should be schema-only).

## Session and request hardening

- **Cookie options in one place**: `HttpOnly: true`, `Secure: true` (behind the ngrok/prod domain),
  `SameSite: Lax`, `MaxAge: 30d`. Today `Options` is rebuilt as `{Path: "/"}` on every write, which
  silently drops everything else.
- **CSRF**: `middleware.CSRFWithConfig` on the closed group, with the token in a hidden field in every
  form (there are ~10 forms; a template helper renders it).
- **Rotate the session on login** (new cookie, old invalidated) to kill session fixation.
- **Rate limit login** (`middleware.RateLimiter` keyed by IP+email, e.g. 5/min) — bcrypt cost 14 already
  makes each attempt ~1s, which is a DoS lever as much as a brute-force brake. Drop to cost 12.
- **Secrets**: fail fast at boot if `SECRET` is missing or shorter than 32 bytes (an empty key makes
  the cookie store forgeable), and stop loading `.env` implicitly in production.
- **CORS**: delete `middleware.CORS()` (same-origin app) or allow-list the real host (B21).

## Admin surface quality

The CRUD pages are functional but unguarded and unlogged. Add:

- **validation on save** for every resource (items: `kind`-dependent fields from plan 02; levels:
  monotonic; recipes: no cycles; users: email format, unique),
- **pagination + search** on `users`, `items`, `inventories` (the tables already overflow at 34 items),
- **soft-delete visibility**: deleted rows vanish but nothing shows they were deleted; add a "show
  deleted" toggle (all models use `gorm.Model`),
- **audit log**: `admin_audit(id, user_id, action, table, row_id, diff_json, created_at)` written by a
  small helper on every create/update/delete. When a player reports "my gold vanished", this is the
  only way to answer,
- **danger confirmations** already exist for delete — extend the pattern to role changes and user
  deletion.

## Acceptance criteria

- Non-admin session → `403` on all eight admin paths, `200` on `/game/*`; a regression test asserts it.
- A fresh database has **no** users until the seed command is run.
- Login with a deleted user and a valid cookie → redirected, not served.
- `curl` cross-origin POST without a CSRF token → rejected.
- Every admin mutation writes exactly one audit row with the before/after values.
