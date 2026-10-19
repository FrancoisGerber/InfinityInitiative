# 03 — Progression and balance

**Fixes:** B9 (NaN/Inf XP bar at the cap), B14 (one level per action), B15 (character XP divisor)
**Effort:** M

## What exists

- `levels` — one table, 99 rows keyed by a numeric `name`, holding `experience_needed` per level, shared
  by every skill *and* the character.
- `character_skills` — per skill `experience_points` + `level_id`.
- `characters.experience_points` — recomputed on every action as `sum(skill XP) / len(Skills)`.

Three structural problems: the curve is global (every skill and the character share thresholds), the
character total is a derived average, and the level-up check is a single `if`.

## Target model

1. **Skill XP advances in a loop.** After adding XP:
   ```go
   for next.ID != 0 && cs.ExperiencePoints >= next.ExperienceNeeded {
       cs.LevelId = next.ID
       next = nextLevel(next.Name + 1)
   }
   ```
   One click can cross several thresholds; the UI shows them stacked ("Mining 12 → 15").
2. **Character XP is its own column**, incremented directly (e.g. 25% of the action's skill XP).
   Adding a skill must not rewrite anyone's level — the current average does exactly that (B15).
3. **Level cap is explicit.** `levels` ends at 99; at the cap render "MAX" with a full bar and skip the
   division entirely (B9). Guard `newMax <= 0` as well, so a missing threshold row can never produce
   `NaN` again.
4. **Thresholds validated on save**: monotonic increasing, level 1 = 0, no duplicates, no gaps —
   a `CHECK`/validation in the admin editor plus a unique index on `name`.

## Anchoring the numbers (research-driven)

The game currently has XP-per-action with no relationship to time, value or risk. The references give
three anchors worth adopting *before* combat ships:

| Anchor | Rule |
|---|---|
| Action XP ↔ runtime | XP per click proportional to the action's cost (inputs consumed), not to a timer. Diamond Crystal (5 Sapphires + 60s) must beat Coal on XP/click, or nobody crafts. |
| Skill → stat | Every skill feeds one primary stat (from `00-research-and-principles.md`): Mining → Strength/Defence, Fishing → Dexterity, Alchemy → Speed … so levelling a trade skill visibly raises combat stats. |
| Class effects | Fighter/Mage/Ranger/Cleric grant flat +% XP in their domain and unlock a talent at Combat 10/35/70. Cheap lookup, gives `characters.class` a job. |

Combat stats themselves (Speed → turn order, Dexterity → accuracy, crit chance/damage) are specified
in `../game-systems/04-battle-pve.md`; this plan only needs the columns and the level→stat curve
(2.4 per level, matching the reference).

## Balance pass (do it once, with data)

- **XP curve**: current thresholds are imported wholesale; sanity-check the first 10 levels against
  the actual XP per action (Coal = 100 XP, 1s runtime) so level 10 arrives in minutes, not hours.
- **Gold**: there is no sink at all today — 34 items, 0 monsters, and gold only appears in the character
  card. Do not add gold rewards (battles) until the Shop plan lands; otherwise the number inflates to
  meaninglessness in a week.
- **Item values**: add a `SellPrice`/`Value` column when the Shop needs it, derived from level + kind,
  not hand-typed per row.
- **Stacking**: `MaxInventorySize = 36` with 34 items means the bag is effectively unbounded for
  resources and tiny for gear. Re-tune alongside the Bank plan (game systems 02).

## Acceptance criteria

- At max level the panel renders "MAX" and no `NaN`/`Inf` appears anywhere in the DOM.
- A single action crossing three thresholds lands on level +3 with the XP remainder preserved.
- Adding an 8th skill leaves every existing character's level unchanged.
- `levels` cannot be saved non-monotonically.
- A spreadsheet (or CSV export) of XP/click, XP/hour of clicking at 1 click/sec, and inputs consumed
  exists for the first 20 items — balance by arithmetic, not by feel.
