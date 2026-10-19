# 05 — Dungeons

**Sidebar:** Game → Combat → Dungeons · **Size:** M · **Depends on:** 04 (combat engine + drops) · **Status:** plan

## Why

A single monster fight is a grind unit. A dungeon is a *run*: a gauntlet of encounters with a
boss at the end, a key you have to earn, and loot you cannot get elsewhere. It is the cheapest way
to turn the combat engine into content with pacing.

## Shape

A dungeon is an ordered list of encounters (3–5) plus a boss, one entry item, a level
requirement, and its own loot table.

```
Enter (consume key) → encounter 1 → 2 → 3 → boss → rewards
                          ↓ HP carries over, flee available between encounters
```

**HP does not reset between encounters.** That is the entire tension: do you push to the boss on
40 HP, or leave with the guaranteed reward?

## Difficulty gate (adopted from IdleMMO)

IdleMMO gives every dungeon a **difficulty rating** and gates entry on total combat stats:

- entry needs combat stats ≥ **70%** of the rating, or the door is shut;
- after a clear you lose health, and *how much* depends on stats vs rating (at exactly 70% you lose
  everything; at 100% about half; at ≥130% almost nothing);
- above 130% you gain a **Magic Find** bonus that scales with how far over you are — the reason to
  come back and farm an easier dungeon.

That single number replaces a level requirement and a loot table tweak in one go. Our version:
`difficulty` on the dungeon row, `combatStats = Attack + Protection + Agility + Accuracy` derived from
the character, and health cost / magic find computed from the ratio on entry.

## Resolution model (the key decision)


- **(A) Resolve the whole run in one request** and render a long log. Fits the blocking engine
  perfectly, zero state, but the player is a spectator — the "should I push on?" decision lives
  only as *before* the run.
- **(B) Step by step**: each encounter is one POST returning the next log + the choice buttons
  (continue / flee / drink potion). State lives in `character_dungeon_runs`. This is where the
  tension actually happens, and HTMX makes it ~free.

**Recommendation: (B).** It is what makes a dungeon different from "fight five monsters in a row",
and the extra state is one row.

## Data

```go
// dungeons - admin authored
Name, Description, Image, Level
EntryItemId   uint   // key item, 0 = free entry
CooldownHours int64  // repeat lockout
FirstClearXp, FirstClearGold int64
RewardItemId, RewardItemAmount uint

// dungeon_encounters
DungeonId, MonsterId uint
Step  int64  // 1..N
IsBoss bool

// character_dungeon_runs
CharacterId, DungeonId uint
Step       int64
Health     int64  // HP carried between encounters
StartedAt  time.Time
Status     string // active | cleared | failed | abandoned
Log        string // accumulated rounds, so the run can be replayed

// dungeon_drops (same shape as monster_drops)
DungeonId, ItemId uint
Chance float64
MinAmount, MaxAmount int64
```

Admin CRUD: `routes/dungeons` copied from `routes/items`, with a nested encounter editor — that
is more UI work than the model, budget for it.

## Routes and views

- `GET /dungeons` — list with level requirement, key status, cooldown remaining.
- `POST /dungeons/:id/enter` — checks level + key, consumes the key, creates the run, resolves
  encounter 1.
- `POST /dungeons/run/:runId/next` — resolve the next encounter.
- `POST /dungeons/run/:runId/flee` — end the run, keep the loot so far (or forfeit it — decide).
- `POST /dungeons/run/:runId/item/:inventoryId` — use a potion/food mid-run.
- `views/dungeons/`, fragments per step; `oob-side-bar` for HP.

## Decisions to make

1. (A) whole-run resolve vs (B) step-by-step (recommended B).
2. **Key items**: craftable keys (ties Dungeons to Crafting) or gold-purchased (ties to Shop)?
   Craftable is more thematic and gives the crafting skills a real purpose.
3. **Loot on flee**: keep what dropped, or forfeit for the guaranteed exit? Forfeit creates the
   push-your-luck tension; keep is kinder. Middle option: keep consumables, forfeit the boss loot.
4. Cooldown: per dungeon (e.g. 4h), daily clear, or none?
5. Do dungeon monsters emit `monster_killed` events (quest progress)? They should — free synergy
   with 03.
6. Level sync: does the dungeon scale to the player, or is it fixed with a hard requirement?
7. Party dungeons: out of scope until multiplayer exists (07).

## Risks

Carrying HP between encounters means a player who dies at the boss loses the whole run — if the
key was expensive, that feels punitive on a blocking request. Show the odds (enemy stats) before
each step, and make fleeing always available.
