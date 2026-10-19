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
agility   = Speed * 2.4        // turn order each round (this is our "Initiative")
accuracy  = Dexterity * 2.4    // vs enemy evasion -> hit chance
protect   = Defence * 2.4      // damage absorbed
hit        = clamp(0.55 + (accuracy - enemy_evasion) / 200, 0.25, 0.95)
damage     = max(1, round((Strength * 2.4 + Attack) - (enemy_protect + enemy_ArmourRating * 0.5)))
crit       = CritChance% chance, +CritDamage flat
```

Two rounds per click (yours, then theirs) so a click is a *trade*, not a single swing — the exchange
is the unit of play. Always at least 1 damage taken per exchange.


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
- `POST /battle/:monsterId/stack` — build or replace the enemy stack (max 3).
- `views/battle/battle.html`, `views/battle/log.html`.

## Health and death

- **Regen from a timestamp, not a tick**: on every read, `regen = elapsedMinutes * regenPerMinute`
  capped at `MaxHealth`; store `LastHealthTick`. No cron, no sleeping request.
- **Food items exist** (`Item.Type = "Food"`) — eating one outside combat heals, consuming a stack.
  That gives the Cooking/Fishing skills a purpose beyond XP.
- **Death**: HP 0 → fight lost, no drops/XP, and either (a) nothing, or (b) a gold penalty /
  respawn timer. Default to something mild in v1; a harsh penalty on a blocking request feels
  awful.

## Clicker combat (not idle timers)

The player is present. One click resolves **one exchange**, or clears one enemy from the stack:

1. Click `Attack` → resolve one exchange between you and the current enemy, append the log line(s).
2. Enemy at 0 HP → loot roll, XP, advance to the next enemy in the stack.
3. Your HP at 0 → fight over, no rewards from the current enemy.

No auto-repeat, no queued fights, no catch-up on load — deliberately the opposite of IdleMMO's 6-hour
idle actions. The fun has to come from the exchange you just clicked, so the log is the product:
round number, hit/miss, damage, crit, killing blow.

Borrowed from IdleMMO because it makes clicking more interesting:

- **Stacks** of up to 3 enemies (IdleMMO allows 6) chosen before the fight — gives the click rhythm a
  goal, not just a single opponent.
- **Scout** — a click that reveals 1–3 enemies for the stack. `Hunting Mastery` raises how many turn
  up; unengaged enemies flee after ~20 in-session clicks, so scouting is not a free stockpile.
- **At least 1 damage taken per fight**, always — no free wins.
- **Enemy scaling** unlocked at level 80 (slider to 150): higher level enemies mean more XP and a
  bigger **Magic Find** multiplier on loot. It is the clicker's answer to "same monster forever".

## Decisions to make

1. Stack size 3 or 6, and does clearing a stack pay a completion bonus?

2. Do we add a real `Initiative` stat (turn order) to characters and monsters? It makes combat
   deeper *and* makes the game's name literal — but it is one more number to balance.
3. Is combat its own skill row (`Combat` icon already exists in `views/assets/img/skills/`) with
   its own XP/level, or does it feed the character level only?
4. Death penalty: none, gold loss, item loss, respawn timer?
5. Level gating: can a level 1 character attack a level 50 monster (and lose), or is the list
   filtered?
6. Do monsters respawn / have per-monster cooldowns, or is the same monster farmable forever?
7. Should fights cost anything (stamina, food)? With one click per exchange, a cost is what stops
   unbounded farming.

## Risks

- **Reward inflation**: this is the first gold/loot *source*. Ship a sink (Shop, 10) in the same
  release or the numbers will be meaningless within a week.
- Balance data live in the DB, so seeding monsters is content work, not code work — plan ~10
  monsters across the first 2 level bands (with the seeded item set) before considering this done.
