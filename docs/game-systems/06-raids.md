# 06 — Raids

**Sidebar:** Game → Combat → Raids · **Size:** L · **Depends on:** 04, and 07 for the group version · **Status:** plan

## Why

A raid is the first system where **other characters matter**: one huge HP pool that many players
(or many members of one Initiative) chip away at, with rewards proportional to contribution. It is
also the natural home for the "idle" pillar — a raid should progress while you are asleep.

## The bootstrap problem

There is no multiplayer yet: one user maps to one character (`GetCharacterForUser`), the chat bar
is static markup, and there is no party model. So the first raid must be enjoyable solo:

- **v1 (solo/scripted boss)**: a 24–48h boss with a large HP pool. Your attacks land on cooldowns;
  the boss HP is a shared row, so when the group version arrives the same maths keeps working.
- **v2 (Initiative raid)**: the pool is fought by an Initiative (07) — see participation rules below.

Design v1 so v2 is a config change, not a rewrite: everything hangs off a `raid_participants` row,
which in solo mode just has one row.

## Mechanics that fit the blocking engine

An attack is not a fight — it is one *round group* resolved instantly:

```
damage_to_boss = sum of 3 rounds of character damage vs boss defence (crits included)
boss_retaliation = damage_to_character unless dodged
```

- Attack gated by `LastAttackAt + CooldownSeconds` (e.g. 30s) — a stored timestamp, no scheduler.
- **Offline catch-up**: on page load, `missed = min(now - LastAttackAt / cooldown, cap)` and the
  missed attacks resolve in one batch with a summary ("while you were away: 42 attacks, 180k
  damage"). Cap the batch (e.g. 100) so a week away does not become a 10-minute request.
- Boss HP decrement must be an atomic `UPDATE raids SET Health = Health - ? WHERE id = ? AND
  Health > 0` — SQLite with concurrent players will otherwise go negative.

## Data

```go
// raids - admin authored
Name, Description, Image, Level
MaxHealth, Health   int64   // Health is the live pool
BossMonsterId       uint    // reuse the monster stat block for retaliation
StartsAt, EndsAt    time.Time
MinLevel            int64
RewardGold, RewardItemId, RewardItemAmount uint  // participation floor

// raid_participants
RaidId, CharacterId uint   // unique
DamageDealt    int64
Attacks        int64
LastAttackAt   time.Time
Claimed        bool

// raid_drops (chance-based, same shape as monster_drops)
```

## Rewards

Contribution tiers, not winner-takes-all: e.g. 1% damage = base reward, 10% = +1 tier, top 3 =
unique item. Everyone who dealt ≥1% gets something, so a solo player is never locked out.

## Routes and views

- `GET /raids` — active raid card (boss art, HP bar, timer, your damage + rank), history of past raids.
- `POST /raids/:id/attack` — one attack group, returns log + OOB HP bar/rank.
- `POST /raids/:id/claim` — after `EndsAt`, pay out.
- Closing a raid (HP 0 or timer end) needs a trigger: resolve lazily on first page read for that
  raid (`if now > EndsAt && Status == "active" -> finalise`). No cron needed.

## Decisions to make

1. **Solo-first or Initiative-only?** (Recommendation: solo v1, group v2 — but that means building
   the same reward tiers twice. Worth it to avoid blocking on 07.)
2. Attack cadence: manual clicks only, or auto-catch-up while offline? (Recommendation: catch-up —
   it is the "idle" in idle MMO.)
3. Does the boss fight back (character can die) or is it a pure DPS race? Retaliation makes
   HP/Food matter; a pure race is simpler and kinder.
4. Duration and cadence: one raid at a time, weekly reset? Overlapping raids with different levels?
5. Loot: chance-based per participant, or a shared pool distributed by damage share?
6. Can a character join multiple raids at once? (No — cap at one active participation.)
7. What stops a whale from soloing a guild raid? (Damage cap per participant per day, or scale HP
   with participant count.)

## Risks

- **Scale**: a group boss's HP has to be tuned against the damage curve, and the damage curve is
  driven by gear that does not exist yet (0 monsters, 34 items). Ship after 04 + 05 have produced
  real character power data.
- Lazy finalisation means a raid can sit "active" with 0 HP until someone opens the page — that is
  fine, but the rewards must be computed from the same finalisation path, not on read.
