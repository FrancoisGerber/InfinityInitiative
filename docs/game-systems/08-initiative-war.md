# 08 — Initiative War PvP

**Sidebar:** Game → Combat → Initiative War PvP · **Size:** L–XL · **Depends on:** 04 (combat engine) + 07 (guilds) · **Status:** plan

## Why

This is the game's endgame and the reason "Initiative" is in the title: two Initiatives commit
their members to a time-boxed war, and the side that wins takes something real. It is also the
place where the word earns its double meaning — battles are resolved in **initiative order**.

## Shape (rebuilt around IdleMMO's Conquest)

IdleMMO does not run live PvP arenas — it runs **assaults on locations**: a guild picks a place, and
every enemy its members defeat there feeds guild EXP, conquest EXP and the member's Guild Mastery.
At the end of a season the top guilds in each location *own* it and get tiered percentage boosts
(primary skill EXP, dungeon magic find, combat EXP …) for the next season.

That is a much better fit for a clicker with one player per account than duel queues, and it keeps the
name honest: the war is fought over territory, and every clash inside it is resolved in **initiative
order** from the Speed/Agility stat.

```
declare war on an Initiative  ──►  war opens
   ├── each member picks a target location and clicks "Assault" (a clash resolves instantly)
   ├── every won clash: +war score, +Initiative EXP, +Guild Mastery to the member
   ├── when two members attack the same location in the same tick, resolve a head-to-head clash
   └── at season end: highest score owns the location, tiered boosts for its members
```

**Duels still exist, but as the exception** — the flavour event when two Initiatives meet in the same
place — not as the primary loop. No live sockets, no lobbies, no countdowns: everything resolves on
the click that triggers it.


## How a duel resolves

Reuse the PvE engine with the attacker's and defender's real stat blocks, ordered by their
`Initiative` stat (see 04, decision 2). Defender gets a small home-terrain bonus (+5% defence) if
they were set to `defend`.

- Attacker pays nothing to attack; attacking breaks their defensive stance for the cooldown.
- Cooldown per character (e.g. 10 minutes), stored as `NextAttackAt`.
- **Every duel in the feed is a click.** No offline catch-up: an assault click resolves your
  character's next clash and writes the feed entry. Members who are not clicking are not fighting.

- Every resolved duel writes a log row the player can read — the war feed is the content.

## Data

```go
// wars
AttackerId, DefenderId uint   // initiative ids
StartsAt, EndsAt time.Time
ScoreAttacker, ScoreDefender int64
Status string                 // pending | active | finished | cancelled
StakeKind string              // gold | item | ranking
StakeAmount int64
WinnerId uint                 // initiative id, 0 = draw

// war_members
WarId uint; CharacterId uint; InitiativeId uint
Stance string                 // attack | defend
NextAttackAt time.Time
Attacks, Wins, Losses int64
DamageDealt, DamageTaken int64

// war_duels  (the feed)
WarId uint; AttackerCharId, DefenderCharId uint
Log string; Winner uint; CreatedAt time.Time
```

## Stakes and anti-abuse

- **Stakes options**: gold from the loser's treasury (clean), an item from the vault (spicy, duping
  risk), or ranking points only (safe). v1 should be ranking + a modest gold transfer.
- **No item loss from your character** — losing gear you equipped in PvE to a PvP duel is the
  fastest way to make players quit an idle game.
- **Matchmaking**: pair Initiatives by level ±2 and member count, not by whoever declares first.
- **Anti-farm**: no war rewards for duels against the same target twice, a minimum participation
  (e.g. ≥3 members and ≥10 attacks on each side), and alt-guild detection (same user id in both
  Initiatives → void the war).
- **War exhaustion**: cooldown before the same pair can fight again, or rivalries become permanent
  griefing.

## Routes and views

- `GET /war` — active war page: score bar, timer, member leaderboard, feed, stance selector,
  attack button with cooldown.
- `POST /war/declare/:initiativeId`, `POST /war/:id/accept`, `POST /war/:id/stance`,
  `POST /war/:id/attack`.
- `GET /wars` — history and standings (this is where reputation lives).
- War card + score bar want a dedicated `oob-war` fragment; the feed is append-only via `hx-swap`.

## Decisions to make

1. **How does a war start?** Declared + accepted by leaders, or automatic weekly pairing (like a
   league)? Automatic is better for an idle game with few players — nobody has to be online to
   accept.
2. **Scoring**: wins only, damage dealt, or a point value per duel weighted by opponent level?
3. **Who can be attacked**: any member, or only members who set `attack` stance (opting out of
   defence)? Opt-in protects casual players and shrinks the griefing surface.
4. **Offline participation**: does a defending member who never logs in get attacked repeatedly?
   Recommend a per-target daily cap rather than invulnerability.
5. **Rewards**: treasury gold, ranking, cosmetic banner, a temporary buff for the winning guild?
6. **Brackets/leagues**: one global ladder or weighted divisions? With 3 characters and no players,
   brackets are premature — but the schema should carry a division field from day one.
7. Is there a **cap on concurrent wars** per Initiative (1)? (Recommendation: 1.)

## Risks

- **This is the most expensive system in the repo**: it needs 04 + 07 + multi-account support and a
  balance pass before it can be tuned. Build it last, and build the feed/log first — a war that
  produces readable stories is worth more than a perfectly balanced ladder.
- Duels against offline opponents are inherently unfair; the daily attack cap is the only cheap
  mitigation. Do not ship without it.
