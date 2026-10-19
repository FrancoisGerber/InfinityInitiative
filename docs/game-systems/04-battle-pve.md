# 04 — Battle PvE

**Sidebar:** Game → Combat → Battle PvE · **Size:** M–L · **Depends on:** nothing technically, but
five other systems depend on *this* · **Status:** plan

## Why this is the keystone

The game has XP, items, equipment bonuses and combat stats (`Attack`, `Strength`, `Defence`,
`ArmourRating`, `Health`, `MaxHealth` on both `characters` and `monsters`) — and **no way to use
any of them**. `monsters` is an admin CRUD table with **0 rows** and no drops. Until combat exists,
gold only ever goes up, items only ever accumulate, and Dungeons / Raids / Wars have nothing to
build on.

## Combat model

Resolve the whole fight in **one request**, server-side, then render the log. The action engine
already blocks while sleeping (`game_controller.go:110`), so sleeping per round would be worse than
useless — do the arithmetic in a loop and render the result.

```
initiative = Attack + (ArmourRating * 0.5)        // who acts first each round
hit        = clamp(0.75 + (atk.Attack - def.Defence) / 100, 0.25, 0.95)
damage     = max(1, round((atk.Attack + atk.Strength * 0.5) - (def.Defence + def.ArmourRating * 0.5)))
crit       = 10% chance, damage * 1.5
```

Loop until one side reaches 0 HP, **hard cap 200 rounds** (draw → loss, HP restored, no rewards).
The log keeps round numbers, damage rolls and the killing blow, so the player can read what
happened and the numbers feel legible — that is the whole point of a log.

## Data

```go
// monsters (extended) - existing: Name, Level, Health/MaxHealth, Attack, Strength, Defence, ArmourRating, ExperienceGained
GoldReward   int64
Initiative   int64   // if we adopt the stat (see decisions)
AttackSpeed  int64   // optional: actions per round

// monster_drops - new, admin authored
MonsterId, ItemId uint
Chance    float64  // 0..1
MinAmount, MaxAmount int64

// characters (extended)
Initiative int64      // turn order
LastHealthTick time.Time  // for regen maths, no background job needed

// optional history
battles (CharacterId, MonsterId, Won bool, Rounds int, Log string, CreatedAt)
```

New admin page under Admin → Database → **Monsters** already exists — extend its editor with the
drops list rather than adding a separate crate of pages.

## Routes and views

- `GET /battle` — monster picker: cards grouped by level band, showing enemy stats, drops and
  "known from" info. Sidebar entry gets `href="/battle"`.
- `POST /battle/:monsterId` — resolve, persist rewards, return a battle-log fragment plus
  `oob-side-bar` + `oob-inventory` (HP, gold, new items all change at once).
- `POST /battle/:monsterId/repeat` — start/stop auto-repeat (see the idle question below).
- `views/battle/battle.html`, `views/battle/log.html`.

## Health and death

- **Regen from a timestamp, not a tick**: on every read, `regen = elapsedMinutes * regenPerMinute`
  capped at `MaxHealth`; store `LastHealthTick`. No cron, no sleeping request.
- **Food items exist** (`Item.Type = "Food"`) — eating one outside combat heals, consuming a stack.
  That gives the Cooking/Fishing skills a purpose beyond XP.
- **Death**: HP 0 → fight lost, no drops/XP, and either (a) nothing, or (b) a gold penalty /
  respawn timer. Default to something mild in v1; a harsh penalty on a blocking request feels
  awful.

## The idle question (decide before building)

"Idle" here can mean three different products:

1. **Active** — the player clicks each fight, sees each log. (Cheapest, matches the current
   action engine exactly.)
2. **Auto-repeat** — a fight resolves, then a cooldown (`NextBattleAt`) gates the next one; on page
   load, resolve the fights that elapsed (batched, capped at e.g. 25) and show a summary. This is
   real idle progress with zero new infrastructure.
3. **Background combat** — server-side scheduler writing to the DB while nobody watches. Needs a
   job runner, which the repo does not have.

**Recommendation: (2).** It is the only flavour that fits the blocking engine, and it is what
Raids/Wars will need anyway.

## Decisions to make

1. Active, auto-repeat, or both (recommended: both, with (2) as the default mode)?
2. Do we add a real `Initiative` stat (turn order) to characters and monsters? It makes combat
   deeper *and* makes the game's name literal — but it is one more number to balance.
3. Is combat its own skill row (`Combat` icon already exists in `views/assets/img/skills/`) with
   its own XP/level, or does it feed the character level only?
4. Death penalty: none, gold loss, item loss, respawn timer?
5. Level gating: can a level 1 character attack a level 50 monster (and lose), or is the list
   filtered?
6. Do monsters respawn / have per-monster cooldowns, or is the same monster farmable forever?
7. Should fights cost anything (stamina, food)? Without a cost, auto-repeat is infinite free XP.

## Risks

- **Reward inflation**: this is the first gold/loot *source*. Ship a sink (Shop, 10) in the same
  release or the numbers will be meaningless within a week.
- Balance data live in the DB, so seeding monsters is content work, not code work — plan ~10
  monsters across the first 2 level bands (with the seeded item set) before considering this done.
