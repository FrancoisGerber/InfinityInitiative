# 10 — Shop

**Sidebar:** Locations → Shop · **Size:** S–M · **Depends on:** a gold source (03/04) · **Status:** plan

## Why

Gold currently has no use at all. The Shop is the **gold sink** — the thing that makes battle and
quest rewards mean something — and the first place a player converts effort into capability
(bank slots, keys, food, potions).

## Two shops or one?

The art folder already carries `gold.png`, `silver.png` and `premium.png`, so the original design
implied more than one currency. Decide now:

- **Gold shop** (this doc): ordinary vendor, gated by character level.
- **Premium shop**: Diamond-priced (Diamonds default to 10 per character, no purchase flow exists).
  If Diamonds never become purchasable, a premium shop is a second progression currency, not a
  monetisation surface — fine, but it means Diamond *sources* have to be designed (dungeon first
  clears, raid tiers, achievements).

Recommendation for v1: **build the gold shop only**, and leave Diamonds as a reward-only currency
with no sink until there is something worth buying.

## What it sells

| Category | Examples | Purpose |
|---|---|---|
| Consumables | food (`Item.Type = "Food"` already exists), potions | make Cooking/Fishing/Alchemy matter |
| Utility | dungeon keys (05), bank slots (02), market listing tokens | sinks that scale with progress |
| Gear | basic weapons/armour at low level only | prevents dead ends; must not outclass crafting |
| Sell | anything in the bag/bank at a fraction of buy price | first way to convert surplus into gold |

The gear line is the dangerous one: if the Shop sells better items than Crafting produces, the
crafting skills stop mattering. Rule: the Shop only sells what a player could craft one tier below
their level, and never endgame gear.

## Data

```go
// shop_items - admin authored
ItemId uint
BuyPrice  int64   // 0 = not for sale in this direction
SellPrice int64   // what the shop pays; 0 = doesn't buy this item
Category  string  // consumable | utility | gear
RequiredLevelId uint
Stock     int64   // -1 = unlimited
RestockSeconds int64
LastRestockAt time.Time
```

Restock is resolved lazily on read (`stock = min(maxStock, stock + elapsed/restockSeconds)`), same
pattern as everything else — no cron.

## Routes and views

- `GET /shop` — tabs by category, item cards with buy/sell price, your gold in the header.
- `POST /shop/buy/:id` — check gold, check bag capacity, debit, add item, `oob`-update gold + bag.
- `POST /shop/sell/:inventoryId` — remove N from the stack, credit `SellPrice * N`.
- `views/shop/shop.html`; sidebar entry gets `href="/shop"`.

## Decisions to make

1. **Pricing model**: hand-written `BuyPrice`/`SellPrice` per item (data-driven, consistent with the
   project) or a formula from `Item.Level` + `Type` (less admin work, harder to tune outliers)?
   Recommendation: formula-generated defaults written into the row by the admin editor, then
   editable — best of both.
2. Sell fraction: 25% of buy price is standard; too high and selling is the primary economy, too
   low and surplus items feel worthless.
3. Stock: unlimited, restocking, or one-off limited stock? (Recommendation: unlimited for
   consumables, restocking for utility, one-off for gear.)
4. Does the Shop **buy everything**, or only `Resource`/`Food` types? Buying gear back creates a
   gear-flipping loop with the Market (09).
5. Buyback of mis-sells — v1 or later? (Nice for trust, cheap to add as a last-transaction row.)
6. Do Diamonds appear in the Shop at all in v1? (Recommendation: no.)

## Risks

The Shop is where an economy becomes visible: the first week of numbers will tell you whether the
battle rewards (04) are inflated. Build the sell price before the buy price — it is the cheaper
half and immediately makes surplus meaningful.
