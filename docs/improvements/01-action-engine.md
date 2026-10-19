# 01 — Action engine

**Fixes:** B3 (non-atomic, non-idempotent), B4 (60-second sleep in the handler), B18 (abandoned
requests lose items), B19 (double payload on error) · **Effort:** M–L

## What happens today

`POST /game/action/:itemId` (`routes/game/game_controller.go:57`) runs five steps in one handler:

1. load item, character, skill progress, recipes;
2. `CharacterHasItems` (read);
3. `RemoveRecipeItemsFromInventory` (write);
4. `time.Sleep(item.Runtime)` — up to **60 seconds**;
5. add XP + character XP + the output item, then render three fragments.

There is no transaction anywhere in that list, no lock, and no request identity. Two clicks = two
payouts; a crash between 3 and 5 = ingredients gone; six concurrent 60-second actions = six blocked
worker goroutines and a SQLite connection storm.

## Target design

**One request = one atomic, idempotent action, resolved in milliseconds.**

```go
// data/models/action.go
type ActionResult struct {
    Skill      CharacterSkill   // new XP / level for the header
    Character  FullCharacter    // gold, level, inventory count
    Gained     []InventoryDelta // +1 Diamond Crystal, -5 Coal, -1 Ruby
    LeveledUp  []string         // "Mining 12 -> 13"
    Elapsed    time.Duration
    Replay     bool             // true when request_id was already applied
}

// RunAction resolves one action. Everything happens inside a single transaction.
func (g *Game) RunAction(characterID, itemID uint, requestID string) (ActionResult, error)
```

Rules:

- **Consume conditionally.** Replace read-then-write with
  `UPDATE inventories SET amount = amount - ? WHERE character_id = ? AND item_id = ? AND amount >= ?`
  and treat `RowsAffected = 0` as "not enough resources". No gap between check and spend.
- **One transaction** around spend → XP → output. Any error rolls back all of it.
- **Idempotency key.** `action_log(id, character_id, request_id UNIQUE, item_id, result_json,
  created_at)`. A repeat of the same `request_id` returns the stored result (`Replay: true`) instead
  of paying out again. The client sends the key in a hidden field; htmx double-fires stop mattering.
- **No sleep.** The action completes now. If the game wants a rhythm, the *client* paces it (a
  disabled button until the local animation ends) or the server returns a short
  `next_action_at` (≤1s) the client honours. Nothing blocks a worker.
- **Return a result, not HTML.** The handler maps `ActionResult` onto fragments (`oob-action`,
  `oob-side-bar`, `oob-inventory`, `oob-alert`). Rendering stays out of the model.

## UX consequence (this is the clicker part)

Right now the "loading bars" spinner covers a frozen request. Once the action is instant, the panel
should feel like a click loop:

- button press → numbers tick immediately (XP bar, item count, gold);
- a compact **action log** line per click ("+1 Oak Log · 25 XP · Mining 12 → 13");
- no page jump, no full reload — the same three OOB swaps, just instant;
- an explicit "out of Coal" error state that does not clear the rest of the panel.

## Migration steps

1. Add `action_log` + `RunAction` beside the existing method; wire the controller to it behind the
   same URL so the UI keeps working.
2. Delete `time.Sleep` and the four inline mutations from the controller.
3. Switch the model's `Update()` calls to explicit column maps (see 06) — otherwise the transaction
   writes associations and defeats the point.
4. Delete the now-unused `CharacterHasItems` / `RemoveRecipeItemsFromInventory` helpers.
5. Add the idempotency key to the two action forms.

## Acceptance criteria

- Two concurrent `POST /game/action/:id` with the same request id → **one** payout, second responds
  with the stored result.
- Killing the request mid-flight (client disconnect) leaves the DB exactly as before the click.
- No handler exceeds ~100 ms; `p99` measured over 100 clicks.
- Running out of an ingredient returns a 4xx fragment and changes nothing.
- `go test -race ./...` clean with 20 concurrent action requests against a temp DB.
