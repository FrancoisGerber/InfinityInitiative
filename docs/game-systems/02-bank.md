# 02 — Bank

**Sidebar:** Information → Bank · **Size:** S–M · **Depends on:** nothing · **Status:** plan

## Why

Inventory is hard-capped at `MaxInventorySize = 36` (`data/models/data.go`) and the action
engine refuses to run when it is full (`game_controller.go:93`). With 34 items and 42 recipes,
a player hits that wall quickly and their only option is destroying resources. A bank is the
pressure valve, and it is a prerequisite for any economy: long-lived stock for quests, dungeons
and guild contributions.

## The one real design decision

**Where do banked items physically live?**

- **(A) `inventories.Location`** — add `Location string` (`"bag"`, `"bank"`) with an index on
  `(character_id, location, item_id)`. One table, one stacking path, one transfer function.
- **(B) separate `bank_items` table** — cleaner separation, but duplicates every query
  (`AddNewItemForCharacter`, `GetInventoryItemForCharacter`, `RemoveRecipeItemsFromInventory`)
  and doubles the code that can drift.

**Recommendation: (A).** Every existing query becomes `... WHERE location = 'bag'`, and the
transfer is a single `UPDATE ... SET location = ?` — the safest way to avoid item duplication,
which is the classic bank bug.

## Data

```go
// inventories (extended)
Location string // "bag" | "bank", default "bag", indexed with character_id + item_id

// characters (extended)
BankSlots int64 // default 60, upgradeable
```

Nothing else. No new table, no new admin page.

## Routes and views

- `GET /bank` — two grids side by side (bag 36 / bank N) reusing the existing storage-grid
  fragment style; a gold row showing carried gold (see decision 3).
- `POST /bank/deposit/:inventoryId`, `POST /bank/withdraw/:inventoryId` — instant, no `Sleep`.
  Return `oob-inventory` + an `oob-bank` fragment so both grids refresh in one swap.
- `POST /bank/deposit-all` / `withdraw-all` for a whole item stack.

## Rules to pin down

- Slot counting must be **distinct item stacks**, matching how the inventory grid counts today —
  the existing full-inventory check counts rows, not amounts.
- Bank is storage, not a shortcut: recipes keep consuming from the bag only, so the player still
  has to plan what to carry. (Alternative: "use from bank" — convenient but it removes the
  inventory-management game entirely.)
- Transfers are free; the cost is slots, bought through the Shop (10).

## Decisions to make

1. (A) vs (B) above — settles the container pattern for House storage too.
2. Does the bank hold **gold** as a vault (protecting it from PvP loss later), or items only?
3. Bank slot count: flat 60, or upgradeable in +10 steps for gold?
4. Can the bank be opened from inside a dungeon/raid (i.e. mid-run withdraw)? Default: no.
5. Does destroying an item from the bank exist, or is bag-only?

## Risks

Item duplication is the failure mode: any transfer path must be one statement inside a
transaction, and the "bag full" check has to run before the withdrawal lands.
