# 07 — Initiatives (guilds)

**Sidebar:** Locations → Initiatives · **Size:** L · **Depends on:** multi-character play (see Risks) · **Status:** plan (concept agreed)

## What it is

**An Initiative is the game's guild** — a player organisation with a shared hall, a shared
treasury, a level, and the thing that fights other Initiatives (08). See `AGENTS.md` for the
settled definition and the evidence (sidebar placement, `Initiative War PvP` under Combat).

The fiction that makes it fit the game rather than being bolted on: an Initiative is an
*undertaking* — a company of adventurers pooling what they gather. Your skills already produce
resources nobody else can use; the Initiative is where surplus becomes power.

## Loop

```
gather / craft  →  contribute items or gold  →  Initiative XP  →  Initiative level
                                                                     ↓
                                            unlocks: member cap, vault slots, perks, war slots
```

Perks are the reason a solo player joins one (and why an empty guild is a real cost):

| Upgrade | Effect | Notes |
|---|---|---|
| Charter | +5 members per level | gate on Initiative level |
| Vault | +N shared storage slots | reuses the Bank container pattern (02) |
| Workshop | +X% skill XP for members | multiplier read at action time |
| War Drums | +war attack slots / faster cooldowns | feeds 08 |
| Bounty Board | +Y% gold from battles | feeds 04 |

**Never bake a perk into a character row.** Compute it per request from
`initiative_members → initiatives → initiative_upgrades`, or buffs will persist after leaving.

## Data

```go
// initiatives
Name, Tag, Description, Image string  // Tag = 3-5 chars, unique
LevelId uint; ExperiencePoints int64
Gold int64                      // treasury
MemberCap int64                 // base 10, raised by Charter
InviteOnly bool
CreatedBy uint                  // character id

// initiative_members
InitiativeId, CharacterId uint   // unique per character -> one initiative each
Rank string                      // leader | officer | member
Contribution int64               // lifetime XP contributed
JoinedAt time.Time

// initiative_upgrades
InitiativeId uint; Key string; Level int64   // Key = "charter" | "vault" | "workshop" | ...

// initiative_log
InitiativeId uint; CharacterId uint; Kind string; Detail string; CreatedAt time.Time
```

## Routes and views

- `GET /initiatives` — directory: name, tag, level, members, open/closed; create form.
- `GET /initiatives/:id` — the hall: banner, roster with ranks and contribution, upgrade tree,
  treasury, contribution log, message board (reuse the chat-bar markup pattern).
- `POST /initiatives` (create), `POST /initiatives/:id/join|leave|kick|promote|contribute|upgrade`.
- All rank checks server-side; `leader` cannot leave without transferring leadership.

## Content to author (data, not code)

Names and flavour are content: seeded example Initiatives ("The Ashen Company", "Ledger of Nine")
make an empty directory look alive on day one, and give the war system something to fight.

## Decisions to make

1. **Creation cost and requirement** — gold + character level + a charter item? Free creation
   means a guild per alt.
2. **One Initiative per character, forever?** Cooldown on re-joining to stop hopping to the winning
   side mid-war.
3. **Member cap** — flat 20, or Charter upgrades? Does the cap count alts and inactive members?
4. **Shared vault** (a real guild bank) in v1, or treasury gold only? The vault is the strongest
   social hook and the strongest duping risk.
5. **Perks: percentage buffs or unlocks?** Percentages compound with everything else and are hard
   to balance; unlocks (extra war slot, extra vault row) are safer.
6. **Rank permissions** — who can spend gold, who invites, who declares war?
7. Does the Initiative own the **House** (11), or is House personal and the Initiative hall is
   separate? (Recommendation: separate — personal housing sells cosmetics later.)
8. Moderation: what does an admin do about a slur in a guild name, given there are no admin tools
   for anything social yet?

## Risks

- **Multiplayer reality**: one user = one character, and no party/invite plumbing exists. A guild of
  10 means 10 accounts. Either accept that (it is a personal project, friends-and-family scale) or
  build multi-character accounts first (a bigger change: `GetCharacterForUser` assumes one).
- Guild banks are the classic dupe vector: every transfer inside one transaction, and log it.
