# 00 — Research and principles

Sources: the IdleMMO wiki (primary), the live game signed in as `@Alir`, the Melvor Idle wiki
(secondary), the OSRS and RS wikis (vocabulary), plus this repo's existing code.
**Nothing here was derived from reading another game's source** — wikis and UI only.

## The rule that overrides everything

**Clicker, not timers.** IdleMMO's core loop is: start an action, close the game, it runs for 6 hours
(8 with membership), come back and collect. We are deliberately building the opposite.

- Progress happens **while the player is present and clicking**. The page is the engine.
- No offline progress, no background jobs, no accumulated-actions catch-up, no "start and leave".
- Cooldowns exist only as anti-spam (seconds, not hours) and must resolve visibly on the click that
  crosses them.
- "Idle" in this game means *low-attention, one-click-repeatable*, in the Melvor "zen" sense — not
  *away from keyboard*.

Every plan below has been amended for this. If a section still hints at timers/catch-up, it is a bug
in the doc.

## IdleMMO system inventory, and what we do with it

| IdleMMO system | What it actually does | Verdict for us |
|---|---|---|
| 6–8h idle actions | Action runs while offline, capped by membership | **Reject** — this is the thing we replace |
| Hunting | `Hunting Mastery` + `Movement Speed` decide enemies found per second; uneaten enemies decay (~10%→2.5% per 4h) | **Adapt** — a "Scout" click reveals 1–3 enemies; unengaged ones flee after a short in-session window |
| Battle stacks | Select up to 6 enemies, auto-resolves through the stack; min 3s per enemy; min 1 damage taken per battle | **Adapt** — queue a stack, one click resolves one exchange (or clears one enemy); you always take at least 1 damage |
| Enemy scaling | Level 80+ can scale enemies to own level; 100+ slider to 150; better Magic Find and EXP | **Adopt** — scaling unlocked by level, boosting loot roll quality and XP |
| Dungeons | No interaction needed, main source of recipes; difficulty rating; entry needs combat stats ≥ 70% of difficulty; health lost after; Magic Find bonus above 130% | **Adopt almost wholesale**, with the run resolved click-by-click instead of over minutes |
| World bosses | Shared HP, queue opens 2h before, 600–900s battles, 8–11h respawn, pet eggs | **Adapt** — shared HP pool and respawn are fine (world state, not player progress); drop the pre-queue timer, keep click-driven attacks |
| Guilds | Level raises member cap and unlocks raids/challenges; 24h re-join cooldown; trade-locked ("banished") guilds | **Adopt** |
| Guild Marks | Guild currency: challenges earn 1, refreshing costs 4, raids cost by difficulty | **Adopt** — replaces a plain gold treasury as the guild's own resource |
| Guild stockpile | Resources contributed for hall upkeep only — **cannot be withdrawn** | **Adopt** — it is the dupe-safe version of a guild bank |
| Guild challenges | 24h, item-contribution objectives, created by ranks, max per day scales with guild level, limited refreshes | **Adopt** — this is our Quests, guild-flavoured |
| Guild raids | Leader schedules, lobby before start, XP scales with participants, leave = forfeit | **Adapt** — no scheduled countdown; the leader starts it and members join while it is open |
| Guild conquest | Assault a location, any kill feeds guild+conquest EXP, top 10 guilds own the territory next season with tiered % boosts | **Adopt as our War** — it is territory conflict, not duels, and it needs no live PvP |
| Guild hall | Guild level 10+, components (conquest banner, teleport beacon, mission table), decay and repairs, marks cost | **Adopt** — this is where Initiative upgrades live |
| Market | Player listings **and** market orders; listings expire after 90 days | **Adopt** |
| Auctions | Valuable item put up for bids | **Adopt as a later tab** of the Market |
| Merchants | Shards (only earnable in-game) for skins/pets/utility; general goods; crest forge | **Adapt** — our Shop, with a second earn-only currency |
| Trading | Direct player-to-player, with "high risk" account flags | **Defer** — needs trust/anti-abuse work |
| Classes | Permanent effects (+XP%, +efficiency) plus battle talents at Combat 10 / 35 / 70 | **Adopt** — we already have a `Class` field and Cleric/Fighter/Mage/Ranger icons |
| Stats | Strength→Attack Power, Defence→Protection, Speed→Agility (battle pace), Dexterity→Accuracy, +Crit Chance/Damage, **2.4 per level**, ascension perks | **Adopt a subset** — see stat mapping below |
| Skills | 20+, including Dungeoneering (unlocks dungeons), Bartering (+% vendor gold, capped 20%), Hunting Mastery, Guild Mastery, Meditation, Construction (houses), Slayer, seasonal masteries | **Adopt the ones with consumers** |
| Membership | Paid longer idle time | **Reject** — it only sells the thing we removed |
| Multiple characters per account | Up to 5, 3 active | **Defer** — ours is one character per user (`GetCharacterForUser`) |

## Stat and class model to adopt

Our `characters` table already has `Attack`, `Strength`, `Defence`, `ArmourRating`, `Health`,
`MaxHealth` — IdleMMO's naming is friendlier and gives each number exactly one job:

| Our field | Becomes | Derived | Used by |
|---|---|---|---|
| `Strength` | Strength | Attack Power (+2.4/lvl) | damage dealt |
| `Defence` | Defence | Protection (+2.4/lvl) | damage absorbed |
| new `Speed` | Speed | Agility (+2.4/lvl) | turn order each round → **this is our Initiative stat** |
| new `Dexterity` | Dexterity | Accuracy (+2.4/lvl) | hit chance |
| `Attack` | keep as a flat gear/heirloom bonus | — | damage |
| `ArmourRating` | keep as flat gear bonus | — | damage reduction |
| new `CritChance`, `CritDamage` | as named | — | combat spice |
| none | — | — | Magic Find (loot quality multiplier, raised by scaling/dungeons) |

Classes grant permanent percentage effects plus talents unlocked at Combat level 10/35/70 — that is
cheap to implement (a lookup table + a bonus function) and gives the four class icons meaning.

## Skill roster proposed

Existing 7 (Mining, Woodcutting, Fishing, Smithing, Crafting, Cooking, Farming) stay. Add only skills
that have a consumer, and note the icon already on disk:

| Skill | Consumer | Icon on disk |
|---|---|---|
| Combat | battles, dungeons, wars | `Combat.png` |
| Dungeoneering | dungeon unlock gate | — (make one) |
| Bartering | Shop sell-price bonus (cap +20%) | — |
| Hunting Mastery | Scout reveal rate | — |
| Guild Mastery | Initiative contribution/war rewards | — |
| Slayer | monster contracts (later) | — |
| Alchemy | potions (Feeding combat) | `Alchemy.png` |
| Fletching | bows/ammo | `Fletching.png` |
| Archery | ranged weapons | `Archery.png` |
| Magic | spells | `Magic.png` |

## UI grammar worth copying from the live game

Observed on `@Alir`: three columns — **left rail** (character card, then sections Account / Party /
Skills, each skill a row with a level badge), **centre** (NPC banner with dialogue, pixel scene,
"Choose a log" picker, then a filterable item list with Lv / EXP / value per row), **right rail**
(skill card: level, XP bar with "175 EXP Needed 35%", Ascension Perks, then Metrics: items gathered,
total experience, time spent).

- Hub pages are just bordered rows (Combat · Slayer · Dungeons · World Bosses) — cheap to build with
  the existing `menu` markup.
- The empty state is explicit: "No Action Assigned — this character is not doing an action right now."
- Top bar carries currency chips.

Our Action / Resources / Storage panels already match this closely. Two additions are worth it: the
**right-rail skill card with metrics**, and **filterable item rows** (All items / Level) instead of
pure grids.

## What the other two wikis add

**Melvor Idle** (secondary): bank slots *and tabs* with search; combat triangle
(melee/ranged/magic accuracy vs evasion, max hit, damage reduction); food with **auto-eat** threshold;
slayer tasks; dungeons as waves then a boss. Take: bank tabs/search (10 slots a tab), accuracy-vs-evasion
hit chance, and food as a combat resource.

**OSRS / RS** (vocabulary): quests with explicit requirements (levels, quest points, items), rewards and
a **quest-point meta currency**; bestiary organised by level band and by slayer category; the Grand
Exchange's offer model (buy/sell offers, collection, price guide, trade limits for new accounts).
Take: quest points, bestiary level bands, and GE-style orders for the Market.

## Amendments applied to the plans

- 04 Battle: click resolves one exchange; hunting is a Scout click; no auto-repeat, no catch-up.
- 05 Dungeons: IdleMMO's 70%-of-difficulty entry gate and post-run health loss adopted; step-by-step clicks.
- 06 Raids: shared HP and respawn adopted, offline catch-up removed, attacks are clicks.
- 07 Initiatives: guild levels → member cap, Marks, non-withdrawable stockpile, challenges, hall at level 10.
- 08 War: rebuilt around IdleMMO's **conquest/assault** model (territory + seasonal ranking) instead of pure duels.
- 09 Market: listings + orders + 90-day expiry + auctions.
- 10 Shop: merchants/shards split from general goods; Bartering as the sell-bonus skill.
- 11 House: Construction as the skill that builds rooms.
