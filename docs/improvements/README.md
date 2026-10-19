# Improvements to what already exists

Plans for the systems that **are** built (skills/actions, items, inventory, crafting, progression,
auth, admin, rendering, data layer). The game-system plans in `../game-systems/` cover what is not
built; this folder covers what is.

Read `../BUGS.md` alongside — each plan names the bug IDs it closes.

| # | Plan | Fixes | Effort |
|---|---|---|---|
| 01 | [Action engine](01-action-engine.md) | B3, B4, B18, B19 | M–L |
| 02 | [Items, inventory, crafting](02-items-inventory-crafting.md) | B10, B11, B12, B20 | M |
| 03 | [Progression and balance](03-progression-and-balance.md) | B9, B14, B15 | M |
| 04 | [Security and admin](04-security-and-admin.md) | B1, B2, B6, B17, B24 | M |
| 05 | [Rendering and UX](05-rendering-and-ux.md) | B16, B19, B22, B29 | S–M |
| 06 | [Data layer and tests](06-data-layer-and-tests.md) | B5, B13, B27, B28 | M |
| 07 | [Character switcher UI](07-character-switcher-ui.md) | — (finishes B7 in the UI) | S |

## Suggested order

1. **04 first (security).** B1 is live: any logged-in account is an admin. It is a middleware and a
   seed-removal, not a refactor, and it gates whether the thing can be put on the internet at all.
2. **01 next.** The action engine is the loop every future system plugs into (quests emit its events,
   battles reuse its transaction shape, raids reuse its click resolution). Fixing it after adding
   five systems means fixing it five times.
3. **02 and 03 together.** Inventory semantics and the XP curve are what the combat plans assume;
   both touch `character_skills` and `inventories`, so do them in one branch.
4. **06 throughout.** The single-DB-handle and transaction helpers are the tools 01–03 need; land them
   early and reuse.
5. **05 last.** Rendering is independent and mostly invisible; do it when the data model has settled.

## The rule everything here serves

From `../../AGENTS.md`: **clicker, not timers.** No handler may sleep, nothing accumulates while the
player is away, and every unit of progress is one request the player caused. Plans that cannot meet
that are wrong, not ambitious.
