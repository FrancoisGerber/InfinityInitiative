# Game systems — design plans

Planning documents for every player-facing sidebar entry that currently has no
implementation. **Nothing here is built.** Each file is a proposal to be discussed,
argued with, and then cut down before code starts.

Status legend: **plan** = written, **agreed** = decided with the owner, **built** = shipped.

| # | Sidebar entry | Section | Size | Depends on | Status |
|---|---|---|---|---|---|
| 01 | Character Sheet | Information | S | — | plan |
| 02 | Bank | Information | S–M | — | plan |
| 03 | Quests | Information | M | shared event hook (03) | plan |
| 04 | Battle PvE | Combat | M–L | combat engine (04) | plan |
| 05 | Dungeons | Combat | M | 04 | plan |
| 06 | Raids | Combat | L | 04, 07 | plan |
| 07 | Initiatives | Locations | L | multi-character play | plan |
| 08 | Initiative War PvP | Combat | L–XL | 04, 07 | plan |
| 09 | Market | Locations | M–L | 10, multiple players | plan |
| 10 | Shop | Locations | S–M | currency sources (03/04) | plan |
| 11 | House | Locations | M | 10 | plan |

## Foundations every system must follow

These are the conventions the codebase already establishes; new features that break
them will look and behave differently from everything else.

1. **Models** — one file per resource in `data/models/`, struct embedding `Base` +
   `gorm.Model`, JSON/form/query tags on every field, a `NewX()` constructor, and query
   helpers as methods. Register it in the `AutoMigration` block in `app.go`.
2. **Controllers** — one package per resource in `routes/<resource>/` with
   `SetupEndpoints(g *echo.Group)`; mount in `routes/router.go` on the **closed** group
   (session auth) unless it is genuinely public.
3. **Ownership checks** — the session gives `c.Get("UserId").(uint)`. Every new endpoint
   must resolve the character *from that user* (`GetCharacterForUser`) and never trust an
   ID from the URL. Bank/quest/guild rows are per-character; a missing check is an
   account-takeover-sized bug in a game with currencies.
4. **Views** — full pages render `index.html` with the layout; in-place mutations return
   HTML fragments and use `hx-swap-oob` (existing fragment names: `oob-action`,
   `oob-side-bar`, `oob-inventory`, `oob-resources`, `oob-action-button`). Sidebar
   highlighting is driven by `ActivePage`; daisyUI 5 renamed the active class to
   `menu-active`.
5. **Content is data, not code.** Items, monsters, recipes, levels and every new system's
   content belong in tables edited through an admin CRUD page (the `routes/items` package
   is the template to copy), not in Go constants.
6. **The action engine is blocking.** `RunActiveAction` does
   `time.Sleep(item.Runtime)` inside the HTTP request and then renders. There is no
   background tick, no queue and no offline progress. Anything time-based (combat rounds,
   raid damage, cooldowns) must therefore be **resolved arithmetically in one request**
   from stored timestamps, never by sleeping per round.
7. **Currency** is `Character.Gold` (soft) and `Character.Diamonds` (premium, no purchase
   flow exists). The "silver" in the UI art is not a currency in code.
8. **Levels** are one global `levels` table (rows 1–99) with `ExperienceNeeded`; skill
   progress lives in `character_skills`, character level in `characters.LevelId`.
9. **Progression content exists, combat content does not**: 7 skills, 34 items, 42 recipes,
   99 levels — and **0 monsters**. Icons for `Combat`, `Magic`, `Archery`, `Fletching`,
   `Alchemy` are already on disk under `views/assets/img/skills/`, so those skills are
   planned but unseeded.

## Suggested build order

```
01 Character Sheet ─┐
02 Bank ────────────┤ (independent, good first PRs)
06 Shop ────────────┘ (needs a gold source to be meaningful)

04 Battle PvE ──► 03 Quests ──► 05 Dungeons ──► 06 Raids
      │                                            ▲
      └──────────► 07 Initiatives ──► 08 War PvP ───┘

10 Shop ──► 09 Market        (Market is worthless before other players exist)
11 House                     (needs Shop prices as its cost anchors)
```

Combat is the keystone: five of the eleven systems are waiting on it, and the game has
XP and items but no way to spend or lose anything.

## Decisions to settle before writing code

1. **Is combat turn-based and logged, or instant-resolved?** (04) This decides the shape of
   dungeons, raids and wars.
2. **Does a "Combat" skill become a real row in `skills`?** Or are combat levels separate?
3. **Idle or active?** Today every action is a blocking POST the player triggers. Do we keep
   that, or add offline resolution? Raids and wars are meaningless without the latter.
4. **Multiplayer reality check**: one user maps to one character
   (`GetCharacterForUser`), the chat bar is static markup, and there is no party model.
   Initiatives, wars, raids and the market all assume other players exist.
5. **One `inventory` table with a `location` column, or a separate `bank_items` table?** (02)
   It sets the pattern for every future container.
6. **Where do gold sinks come from?** Without them, battle rewards inflate the economy before
   Shop/Market can absorb it.
