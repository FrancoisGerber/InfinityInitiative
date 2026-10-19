# 01 — Character Sheet

**Sidebar:** Information → Character Sheet · **Size:** S · **Depends on:** nothing · **Status:** plan

## Why

The sidebar already shows a miniature version (`character-info` template inside the menu), and
`/characters` is an admin CRUD table. There is no page a *player* can open to see who their
character is: level, HP, combat stats, equipment, all skill levels in one place.

## Scope for v1

Read-only. No actions, no spending, no mutations — which is why it is a good first PR.

1. Header card: name, class, level + XP bar to next level, HP/MaxHealth, gold, diamonds.
2. Combat block: Attack, Strength, Defence, Armour Rating (aggregate, as
   `GetFullCharacterForUser` already folds equipped-item bonuses in).
3. Equipment: the `inventory` rows with `Equipped = true`, with item art and stats; empty slots
   shown as placeholders.
4. Skills table: per-skill level, XP into current level, XP to next, actions available.
5. Bio / image.

Mirror the reference UI (IdleMMO's right rail): a **skill card** per skill showing level, an XP bar
with "X EXP needed — Y%" and lifetime **metrics** (items gathered, total experience, time spent) is
the single most convincing "this is a real game" element in that UI, and every number already exists
in `character_skills` + `inventories`.


## Data

Nothing new. `Character`, `CharacterSkill`, `Inventory` + `Item` already carry everything.
No migration, no admin page.

## Routes and views

- `GET /character` on the closed group.
- `views/character/sheet.html` rendered through the normal layout; sidebar entry gets
  `href="/character"` and `ActivePage` highlighting.
- Pure server-render, no HTMX needed.

## Decisions to make

1. **Overlap with "Profile"** in the navbar (`href="/characters"`, currently the admin list).
   Should Profile point here and the admin table move under Admin → Characters only?
2. Do we show **planned-but-unused** stats now (HP is default 10, combat has no engine yet), or
   hide the combat block until 04 ships?
3. Optional **combat power score** (one number derived from the four stats) as a quick
   comparison — worth it, or false precision?
4. Does the sheet show other players' characters later (public profiles)? If yes, the route
   shape matters now: `/character` vs `/character/:id`.

## Risks

None technical. The only failure mode is scope creep into "character management" (rename,
respec, class change), which belongs to a separate system.
