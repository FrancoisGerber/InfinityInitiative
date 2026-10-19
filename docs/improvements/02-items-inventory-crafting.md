# 02 — Items, inventory and crafting

**Fixes:** B10 (row-count capacity), B11 (arbitrary row lookup), B12 (equip mutates stacks), B20
(zero-runtime actions, equipment leaking into action lists) · **Effort:** M

## Inventory model

Today an `inventories` row is "some amount of an item, possibly flagged equipped" — and the equipment
flag lives on the same row as the quantity. That one decision causes B11 and B12.

**Target:**

```go
// inventories: exactly one row per (character, item, location)
CharacterId, ItemId, Location string // "bag" | "bank" (bank arrives with plan 02 of the game systems)
Amount         int64
EquippedAmount int64                  // 0..Amount, not a boolean
// UNIQUE(character_id, item_id, location)
```

- **Capacity = distinct stacks** in `location = 'bag'`: `SELECT count(*) FROM inventories WHERE
  character_id = ? AND location='bag' AND amount > 0`. Enforced inside the same transaction that
  inserts, so a full bag cannot be over-filled and a stackable output is never rejected (B10).
- **Equip = move quantity**, not a flip: equipping 1 of 5 sets `equipped_amount = 1`; the grid renders
  "5 (1 equipped)". Unequip reverses it. Zero-amount rows are deleted on read.
- **Slots are validated**: one weapon, one armour, one accessory (extend `Item.Type` into typed
  constants, not free-text). Equipping a second weapon either swaps or refuses — pick one and say so
  in the UI.

## Item semantics

`Item.Type` is a free-text string (`Weapon/Armor/Accessory/Resource/Food`) and `SkillId` decides
whether an item shows up as an action (B20). Separate the two concepts:

- `type` — what it is (resource, food, weapon, armour, accessory, key, currency-token).
- `kind` — how it is used: **action** (has runtime + XP + level requirement), **equipment** (has
  stats), **consumable** (has an effect), **material** (recipe input only).

Validation on save: `kind = action` requires `runtime > 0` and `experience_gained > 0`;
`kind = equipment` requires at least one stat; equipment may never appear in a skill's action list.
This closes the "free XP + free item with runtime 0" hole and stops a mis-entered sword becoming a
gathering node.

## Crafting

- **Recipe integrity**: `UNIQUE(parent_item_id, item_id)`, `amount >= 1`, parent ≠ ingredient, no
  cycles (validate on save).
- **Quantities**: craft ×1 / ×5 / ×max. `×max` is computed as
  `min(floor(amount_i / recipe_amount_i))` and capped by bag space — one click, one transaction, N
  outputs. That is the clicker version of "make-all".
- **Recipe visibility**: show what you can afford vs what you cannot (greyed), with the missing
  ingredient named. The current error string ("You do not have enough resources to make this Item!")
  is the only feedback and it does not say which item.
- **XP**: crafting XP comes from the parent item's `experience_gained`, same as gathering — keep one
  code path.

## Admin editor

The items editor needs guardrails, because it is where every number in the game is authored:

- required fields by `kind` (above);
- warnings for stats wildly outside the current tier (e.g. a level-1 item with 500 attack);
- a "preview as player" panel showing what the row looks like in the action list;
- recipe editor that refuses cycles and self-reference before saving.

## Acceptance criteria

- A full bag accepts a stackable output; a full bag rejects a new distinct item with a clear message.
- Equipping/unequipping 20 times leaves `sum(amount)` unchanged and no zero-amount rows.
- Two weapons cannot be equipped simultaneously.
- An item with `runtime = 0` cannot be saved as an action.
- Craft ×max of a recipe with 3 ingredients yields exactly `min(floor(...))` outputs, atomically.
