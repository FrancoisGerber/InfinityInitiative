# 11 — House

**Sidebar:** Locations → House · **Size:** M · **Depends on:** 10 (prices/costs) · **Status:** plan

## Why

The House is the personal, cosmetic-but-useful space: rooms you buy and upgrade with gold and
crafted materials, each granting a small passive effect. It is the endgame gold sink and the
reason `Crafting` matters once gear is capped.

It is **not** the Initiative hall (07). Keep them separate: personal space is where you show
progress, the Initiative hall is where you cooperate. Merging them means the first thing a new
player does is decorate a guild house they do not control.

## Loop

```
buy room (gold + materials) → place in a slot → upgrade (more gold + materials)
        ↓                                              ↓
   passive buff active                          visual tier changes
```

## Rooms and buffs (v1 candidate list)

| Room | Buff | Why it is safe |
|---|---|---|
| Workshop | −X% action `Runtime` | the action engine already stores runtime per item |
| Library | +X% skill XP | multiplies an existing curve |
| Vault | +N bank slots (02) | an unlock, not a multiplier |
| Kitchen | food heals more / cheaper | gives Cooking a sink |
| Shrine | +X max HP regen per minute | touches 04's regen maths |
| Trophy Room | cosmetic only | prestige |

Rule: **unlocks over multipliers**, and any multiplier must be a single clearly-capped percentage —
these stack with Initiative perks (07) and gear, and three stacked multipliers is how idle games
break their own balance.

**Construction is the skill that builds the house** (IdleMMO calls it exactly that). Room
construction and upgrades should award Construction XP, and the room list should be gated by both
house level and Construction level — that gives an eighth skill a consumer without inventing a new
system.

**No decay, no upkeep.** IdleMMO's guild hall components decay and need repairs; for a personal house
in a clicker that is a chore with no drama. Rooms, once built, stay built.


## Data

```go
// houses
CharacterId uint          // unique, one house per character
Level int64               // gates room count and upgrade caps
Theme string              // visual set later
Slots int64               // default 3, +1 per house level

// room_types - admin authored
Name, Description, Image string
CostGold int64
CostItemId uint; CostItemAmount int64
BuffKind string           // runtime | xp | bank_slots | heal | regen
BuffValue float64
MaxLevel int64
RequiredHouseLevel int64

// house_rooms
HouseId, RoomTypeId uint
Level int64
Slot  int64               // 0..(Slots-1); unique per house
```

## Routes and views

- `GET /house` — house card (level, XP to next, slots), room grid with empty-slot placeholders,
  room detail panel with upgrade cost.
- `POST /house/build/:roomTypeId` (into a chosen empty slot), `POST /house/rooms/:id/upgrade`,
  `POST /house/rooms/:id/move`, `POST /house/theme`.
- Fragments: `oob-house` (grid) + `oob-side-bar` (if HP/regen is affected).
- Buffs are computed **per request**: `house.Buff(characterID, kind) float64`, cached in the handler,
  never written into the character row.

## Decisions to make

1. **Grid editor vs slot list.** A draggable/tile map is a whole UI project; a fixed list of N slots
   with a room picker is a day. Recommendation: slots in v1, grid only if the cosmetic side becomes
   the point.
2. **House level**: separate XP track (upgrades grant house XP), or gate purely on character level?
   A separate track gives the House its own progression, which is the reason players keep spending.
3. Can other players **visit** a house (public URL `/house/:characterId`)? Cheap once Houses exist,
   and it makes decoration meaningful — but it exposes character names, so it is a privacy decision.
4. Do rooms **decay** or need upkeep? (Recommendation: no — upkeep in an idle game is a chore with
   no drama.)
5. Can rooms be **refunded** or moved freely? Moving should be free; refunds should not (otherwise
   it is a free respec merry-go-round).
6. Do buffs apply to every click? They must — computed on the same request that resolves the action.
   A buff that only applies sometimes is a bug, not a feature.

## Risks

- The buff list is the balance surface. Ship 3 rooms, not 6, and measure before adding multipliers.
- The House shares the storage concept with the Bank (02): if the Vault room unlocks bank slots, the
  slot maths must live in one place (`Bank.SlotsFor(characterID)`) rather than being computed in
  three handlers.
