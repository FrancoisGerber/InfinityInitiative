# 09 — Market

**Sidebar:** Locations → Market · **Size:** M–L · **Depends on:** 10 (Shop, for price anchors) + other players · **Status:** plan

## Why

The Market is player-to-player: list an item for gold, someone buys it, you get paid minus tax.
It gives surplus resources a destination and makes specialised skills worth levelling (a miner can
sell ore to a smith instead of grinding both).

## The bootstrap problem

A market with one active player is an empty page. Options:

- **Launch after the game has players** (recommended — this is why it is last in the build order),
  or
- **Seed it with NPC listings** so browsing works from day one, with real player listings layered
  on top. Seeding teaches the UI cheaply and lets the Shop (10) act as a price anchor.

Either way, the schema and UI are the same; only the seeding differs.

## Rules that keep it honest

- **Listing fee** (1% of asking price, minimum 1 gold) — paid on listing, not on sale. This is the
  anti-spam and the gold sink.
- **Sales tax** (5%) on sale — the second gold sink.
- **Expiry** (7 days) — item returns to the bag/bank on expiry, no refund of the listing fee.
- **Price bounds**: min = the item's Shop sell price, max = 50× Shop buy price (or a per-item cap).
  Without a floor, two alts launder gold; without a ceiling, a rare goes for 10^9.
- **Max 5 active listings** per character.
- Item binding: loot marked bound (dungeon first-clears, quest rewards) cannot be listed. Add a
  `Bound bool` to `inventory` when first needed, or v1 simply allows everything.

## Data

```go
// market_listings
SellerCharacterId, ItemId uint
Amount int64
PriceEach int64          // per unit
CreatedAt, ExpiresAt time.Time
Status string            // active | sold | expired | cancelled
BuyerCharacterId *uint
SoldAt *time.Time

// optional, for a "sold recently" price hint
market_sales (ItemId, Amount, PriceEach, SoldAt)
```

## Routes and views

- `GET /market` — browse: filter by type/level, sort by price, paginated; "your listings" panel.
- `POST /market/list` — from an inventory/bank item: amount + price, charges the fee.
- `POST /market/:id/buy` — **one transaction**: claim the listing (`WHERE status='active'`),
  transfer gold, transfer the item (respecting bag capacity — check first!), credit the seller.
- `POST /market/:id/cancel`, and an expiry sweep resolved lazily on read (like raids) — no cron.
- `views/market/`, listing cards reuse the item-grid markup.

## Concurrency and duping

The buy path is the highest-risk code in the project:

1. `UPDATE market_listings SET status='sold', buyer_character_id=? WHERE id=? AND status='active'`
   → if 0 rows affected, someone else bought it (show "already sold").
2. Then, in the same transaction: debit buyer, credit seller, move item.
3. Bag-full must be checked **before** step 1, and the item must go somewhere deterministic if the
   bag fills in between (bank, or refund the purchase).

## Decisions to make

1. Seed with NPC listings, or ship empty and accept a dead page until there are players?
2. Gold only, or allow Diamond-priced listings? (Diamonds have no purchase flow; keeping the
   Market gold-only avoids a de-facto RMT surface.)
3. Sell **units** or **stacks**? Per-unit pricing with a quantity field is more flexible; stacks are
   less clicking.
4. Auction (bids) vs fixed price? Fixed price is a day of work; auctions are a system.
5. Scope: global market, or per-Initiative trading post first (fewer players needed to be useful)?
6. Do listings show the seller's name (trust/reputation) or are they anonymous (no harassment)?
7. Price history graph — nice, but it is a second feature.

## Risks

Two-account self-trading (laundering gold between your own characters to dodge the tax) is
unavoidable without account-level checks; the price floor plus tax makes it unprofitable rather
than impossible, which is enough.
