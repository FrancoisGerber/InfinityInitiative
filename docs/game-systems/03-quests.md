# 03 — Quests

**Sidebar:** Information → Quests · **Size:** M · **Depends on:** shared event hook (built here) · **Status:** plan

## Why

Nothing in the game currently *asks* anything of the player. Skills gain XP forever and items
pile up, but there is no goal, no reward for a session, and no reason to touch a skill you
would not otherwise grind. Quests are also the cheapest way to teach the systems as they ship
(a quest that says "craft 5 Bronze Daggers" is a tutorial that does not feel like one).

## The architectural piece (do this first)

Quests only work if the game emits "things happened". Add one function and call it from the two
places that already mutate progress:

```go
// data/models/quest.go
type QuestEvent struct {
    Kind     string // "item_gained" | "item_crafted" | "monster_killed" | "skill_level" | "gold_earned" | "item_equipped"
    TargetId uint   // item / monster / skill id, 0 when unused
    Amount   int64
}

func (q *Quest) Advance(characterId uint, e QuestEvent) // bumps progress, completes quests
```

Call sites: `game.IncreaseCharacterSkill` / `AddNewItemForCharacter` / `RemoveRecipeItemsFromInventory`
(in `routes/game/game_controller.go`), and later the battle resolver. One hook, every future
feature (dungeons, initiatives) can emit the same events.

## Borrowed structure

- **From OSRS**: quests have explicit *requirements* (skill levels, quest points, items), explicit
  *rewards* (XP, items, unlocks) and pay **quest points** — a meta-currency that gates later content.
  Steal that: `quests.RequiredLevel`, `RequiredQuestPoints`, `QuestPoints` on the character, and
  chains that unlock systems (Dungeons, Initiatives).
- **From IdleMMO**: its *guild challenges* are essentially this system done co-operatively — 24-hour
  objectives ("contribute 50 Oak Logs"), created by officers, contributing pays the member 2 XP per
  item whether or not the goal is met, and the guild earns Marks on completion. Same event hook, one
  extra table (`challenge_contributions`), so a guild-quest board is a cheap second surface once the
  personal quest list exists.

## Data model

```go
// quests - admin authored
Name, Description, Image, Level          // requirement to accept
Kind      string   // gather | craft | kill | skill_level | gold
TargetId  uint     // item / monster / skill
Amount    int64    // how many
RewardXp, RewardGold int64
RewardItemId, RewardItemAmount uint
PrereqQuestId uint  // linear chain, 0 = none
Repeatable bool     // dailies later

// character_quests - per character state
CharacterId, QuestId   uint   // unique index
Progress  int64
Status    string       // locked | active | complete | claimed
StartedAt, CompletedAt time.Time
```

## Routes and views

- `GET /quests` — one page, three groups: Active (with progress bars), Available, Completed.
- `POST /quests/:id/accept`, `POST /quests/:id/claim` (claim = pay out, mark claimed).
- `views/quests/quests.html`, fragments per card; progress bars reuse the existing XP-bar markup.
- Sidebar entry gets `href="/quests"`; a small badge with the count of claimable quests is the
  cheap hook that makes it feel alive.

## Rules to pin down

- **Retroactive credit**: a character already at Mining 20 receiving "reach Mining 10" — auto
  complete, or start at zero? (Recommendation: evaluate `skill_level` and `gold` quests against
  current state on accept; count only forward for gather/craft/kill.)
- **Quest board size**: endless list, or 3 concurrent slots to force choice?
- **Chains over boards**: `PrereqQuestId` gives an ordered story; a board of parallel dailies is a
  different product. Pick one for v1.
- Do quests **gate systems** (e.g. "finish quest 3 to unlock Dungeons")? If yes, the quest hook
  also becomes the feature-flag mechanism.

## Decisions to make

1. Chains vs parallel board vs both.
2. Are quests character-bound or account-bound? (Characters are per-user already; assume character.)
3. Do gather/craft quests **consume** the items, or only count them? Consuming makes them sinks;
   counting makes them free progress. This is an economy lever, decide deliberately.
4. Daily/weekly repeatables — v1 or later?
5. Reward shape: flat XP, or XP scaled to the quest level like items do?

## Risks

The event hook is easy; the failure mode is **double counting** (crafting fires both
`item_crafted` and `item_gained`) and progress lost to a crash mid-transaction. Keep
`Advance` inside the same DB transaction as the action that triggered it, or accept the drift
explicitly.
