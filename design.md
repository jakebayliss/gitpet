# `pet` — Terminal Pet

> A persistent pet companion for your terminal. Hatch a pet, level it up by committing code, evolve it, and maybe get a rare drop.

---

## Concept

You start with one pet. Write code, commit it, your pet gains XP. Level up, evolve, get drops. Small focused commits are rewarded more than giant dumps. Everything is local, single binary, no server.

---

## Project Name

**`pet`** — one binary does everything.

---

## Repository Structure

```
petd/
├── cmd/
│   └── pet/           # Single CLI entrypoint
│       └── main.go
├── internal/
│   ├── gen/           # Pet generation (species, rarity, stats)
│   ├── store/         # SQLite persistence
│   ├── xp/            # XP calculation from git diffs
│   ├── drop/          # Drop table engine
│   ├── evolution/     # Evolution rules, levelling, prestige
│   └── ui/            # Terminal rendering (plain output)
├── config/
│   └── drops.yaml     # Drop tables
└── README.md
```

---

## How It Works

```
You write code and commit it
        |
post-commit git hook runs `pet commit`
        |
`pet commit` reads the git diff
        |
Calculates XP from lines changed, files touched, commit size
        |
Grants XP to active pet, rolls for drop
        |
Prints level-up or drop inline
        |
Done. No daemon. No wrappers. One binary.
```

---

## Commands

```
pet hatch              Hatch a new pet (prompts for name)
pet show               Show your active pet's card
pet init               Install post-commit hook in current repo
pet commit             Called by git hook — calculates XP from last commit
pet watch              Live pet window (Phase 2)
pet list               Show all your pets
pet switch <name>      Change active pet
pet kill <name>        Permanently delete a pet (with confirmation)
pet evolve             Evolve active pet if eligible
pet prestige           Reset to level 1 for a prestige star (requires level 99)
pet log                Show recent commit history with XP
```

---

## Git Integration

### Setup

```bash
# Install hook in a repo (creates .git/hooks/post-commit)
cd my-project
pet init

# Or install globally for all repos
pet init --global
```

`pet init` writes a `post-commit` hook that calls `pet commit`. Global mode uses git's `core.hooksPath` so every repo is covered automatically.

### How `pet commit` works

Reads the last commit's diff via `git diff HEAD~1 HEAD --stat` and calculates XP:

```
lines_changed  = insertions + deletions
files_touched  = number of files in diff

base_xp        = lines_changed (capped at 200 for full value)
file_bonus     = files_touched * 5 (capped at 10 files)
size_penalty   = 0.5x if lines_changed > 500

XP = (base_xp + file_bonus) * size_penalty * bonuses
```

**Bonuses (stack multiplicatively):**
| Bonus | Condition | Multiplier |
|-------|-----------|------------|
| Branch commit | Not on `main` or `master` | 1.25x |
| Streak | Consecutive days with commits (day 3+) | 1.1x per day, cap at 1.5x |

A commit on a feature branch on a 5-day streak:
`XP * 1.25 * 1.3 = XP * 1.625`

**Design philosophy:** small, focused commits on feature branches earn the most XP. Rewards good git hygiene and daily consistency.

---

## Core Systems

### 1. Pet Generation (`internal/gen`)

Your first pet is always **Common** — everyone starts equal. Species is rolled randomly from the common pool. You name it yourself.

Additional pets come from egg drops. Rare eggs pull from the rare pool, legendary eggs from the legendary pool. Rarity is purely a flex label and visual flair — no gameplay advantage.

**What gets rolled on hatch:**
- Species (random from the relevant pool)
- Shiny flag (1% chance)
- Base stats (5 dimensions, floor set by pool tier)

**What the user provides:**
- Name (prompted during `pet hatch`)

**Stat floors by tier:**
| Tier      | Stat floor |
|-----------|------------|
| Common    | 20         |
| Rare      | 50         |
| Legendary | 80         |

---

### 2. Species

Species is **purely cosmetic** — it only determines ASCII art and flavor text. No gameplay advantage. The fun is what you get, not what's optimal.

Three pools, each tied to how you obtain the pet:

**Common Pool** (starter pet)
| Species        | Vibe                        |
|----------------|-----------------------------|
| Axolotl        | Friendly aquatic weirdo      |
| Capybara       | Chill vibes                  |
| Goblin         | Chaotic gremlin energy       |
| Pangolin       | Armored and shy              |
| Tardigrade     | Indestructible micro-beast   |
| Blobfish       | Sad but lovable              |
| Mantis Shrimp  | Punchy and colorful          |
| Nudibranch     | Psychedelic sea slug         |

**Rare Pool** (rare egg drop, ~1 in 250 commits)
| Species        | Vibe                        |
|----------------|-----------------------------|
| Salamander     | Fiery and quick              |
| Dire Wolf      | Loyal, intimidating          |
| Thunderbird    | Crackling storm bird         |
| Crystal Fox    | Translucent, glittering      |
| Chimera        | Three heads, three opinions  |
| Shadow Lynx    | Barely visible, stalks your cursor |

**Legendary Pool** (legendary egg drop, ~1 in 1000 commits)
| Species           | Vibe                           |
|-------------------|--------------------------------|
| Phoenix           | Born from deleted code          |
| Leviathan         | Ancient deep-sea terror         |
| Basilisk          | Turns bugs to stone             |
| Void Cat          | Glitchy, shouldn't exist        |
| Kraken            | Tentacles everywhere            |
| Celestial Beetle  | Cosmic shimmering shell         |
| Mimic             | Looks like a file, actually alive |
| Wyrm              | Coils around your terminal      |

Each species has ASCII art for each evolution stage (4 stages).

---

### 3. Stats

Five stats, each shaped by commit patterns:

| Stat        | Grows when...                                    |
|-------------|--------------------------------------------------|
| **Wit**     | Small commits, few lines, surgical changes       |
| **Depth**   | Large meaningful commits, many files             |
| **Stamina** | Total commit count (raw volume)                  |
| **Luck**    | Random variance — affects drop rates             |
| **Attune**  | Commit streaks (consecutive days with commits)   |

Stats start from base values at hatch and grow via XP allocations on level-up.

---

### 4. XP and Levelling (`internal/xp`)

XP is earned per commit. Level thresholds follow a curve (each level costs progressively more).

Level cap: **99**. Each level costs progressively more XP.

On level-up, the player gets **2 stat points** to allocate.

---

### 5. Prestige (`internal/evolution`)

Hit level 99 and you can prestige. Resets your pet to level 1 but:
- Permanent star added to pet name display (one per prestige)
- Small base stat bonus (+2 to all stats per prestige)
- Prestige count displayed on pet card

No prestige cap. Flex as hard as you want.

```
pet prestige

Are you sure? Gribble will reset to Level 1.
Prestige stars earned: *
Base stat bonus: +2 to all stats

[y/N] > y

Gribble has been reborn! * Prestige 1 *
```

---

### 6. Drop System (`internal/drop`)

One drop roll per commit (out of 1000):

```yaml
table:
  - item: nothing
    weight: 907
  - item: xp_crystal
    weight: 60
  - item: title_scroll
    weight: 20
  - item: evolution_stone
    weight: 8
  - item: rare_egg
    weight: 4
  - item: legendary_egg
    weight: 1
```

**Items:**
- `xp_crystal` — bonus 200 XP to active pet (~1 in 17 commits)
- `title_scroll` — random title for your pet, e.g. "the Relentless", "Bug Slayer" (~1 in 50)
- `evolution_stone` — required to evolve alongside level threshold (~1 in 125)
- `rare_egg` — hatch a pet from the rare pool (~1 in 250)
- `legendary_egg` — hatch a pet from the legendary pool (~1 in 1000)

---

### 7. Evolution (`internal/evolution`)

Requires level threshold **and** an evolution stone.

| Stage | Level Required | Visual              | Stat multiplier |
|-------|---------------|----------------------|-----------------|
| 1     | --            | Small ASCII sprite    | 1.0x            |
| 2     | 30            | Medium ASCII sprite   | 1.3x            |
| 3     | 60            | Large ASCII sprite    | 1.6x            |
| 4     | 90 + rare item| Full-width ASCII art  | 2.0x            |

---

### 8. Terminal Output (`internal/ui`)

Plain terminal output for v1. No TUI framework needed.

**`pet show`:**
```
+======================================+
|  * Gribble  -  Axolotl  -  RARE     |
|  Prestige 2  **                      |
|                                      |
|       /\_/\                          |
|      ( o.o )   * SHINY               |
|       > ^ <                          |
|                                      |
|  LVL 24 ========----  1840/2400 XP   |
|                                      |
|  WIT     ========--  42              |
|  DEPTH   ======----  31              |
|  STAMINA =========- 48               |
|  LUCK    ===-------  17              |
|  ATTUNE  =======---  38              |
+======================================+
```

**`pet hatch`:**
```
Rolling...

You hatched a RARE Axolotl!

     /\_/\
    ( o.o )   * SHINY!
     > ^ <

What will you name it? > Gribble

Welcome, Gribble!
```

**`pet commit` (inline after a git commit):**
```
Gribble gained 87 XP! (1840/2400)
Commit streak: 3 today (1.5x bonus!)
Drop: nothing
```

---

## Storage

Single SQLite database at `~/.petd/pets.db`.

```sql
CREATE TABLE pets (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    species     TEXT NOT NULL,
    rarity      TEXT NOT NULL,
    shiny       INTEGER DEFAULT 0,
    level       INTEGER DEFAULT 1,
    xp          INTEGER DEFAULT 0,
    evolution   INTEGER DEFAULT 1,
    prestige    INTEGER DEFAULT 0,
    stats_json  TEXT,
    active      INTEGER DEFAULT 0,
    hatched_at  DATETIME,
    updated_at  DATETIME
);

CREATE TABLE commits (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    pet_id          TEXT REFERENCES pets(id),
    repo            TEXT,
    lines_changed   INTEGER,
    files_touched   INTEGER,
    xp_earned       INTEGER,
    drop_item       TEXT,
    created_at      DATETIME
);

CREATE TABLE inventory (
    item        TEXT PRIMARY KEY,
    quantity    INTEGER DEFAULT 1
);
```

---

## `pet watch` — Live Pet Window (Phase 2, design TBD)

A persistent terminal window you keep floating while you work. Shows your pet live, reacts to XP/level changes from commits in other terminals.

```
+==================================+
|  Gribble  -  Axolotl  -  RARE   |
|                                  |
|     /\_/\                        |
|    ( o.o )   * SHINY             |
|     > ^ <                        |
|                                  |
|  LVL 24 ========----  1840/2400  |
+----------------------------------+
|  Gribble: *wiggles happily*      |
|  You: how are you doing buddy    |
|  Gribble: i ate a bug. it was    |
|           crunchy. good day.     |
|                                  |
|  > _                             |
+==================================+
```

**Implementation:** Bubbletea TUI. Polls SQLite on interval to detect changes from other terminals.

### Chat options (decide after Phase 1)

| Option | How it works | Pros | Cons |
|--------|-------------|------|------|
| **Canned responses** | Random flavor text per species. No network calls. | Zero cost, no API key, works offline, instant | Gets repetitive |
| **LLM chat** | Pet responds in character via API call. | Feels alive, thematically fits | Requires API key, costs money, latency |
| **Hybrid** | Canned by default. LLM if API key configured. | Best of both | More code to maintain |

Other considerations:
- Idle animations (ASCII art shifts periodically)
- React to events (level-up celebration, drop, evolution prompt)
- Show feed of recent commits

---

## Milestones

| Phase | Scope |
|-------|-------|
| 1 -- Core | `pet hatch`, `pet show`, `pet init`, `pet commit`, generation, SQLite store, ASCII art, XP from git |
| 2 -- Watch | `pet watch` live window with Bubbletea (chat approach TBD) |
| 3 -- Drops | Drop table engine, inventory, `pet_egg` hatch flow |
| 4 -- Evolution | Evolution rules, stage sprites, `pet evolve` |
| 5 -- Prestige | `pet prestige`, prestige stars, stat bonuses |
| 6 -- Multi-pet | `pet list`, `pet switch`, `pet kill` |
| 7 -- Polish | `pet log`, colours, shiny rendering, idle animations |

---

## Non-Goals (for now)

- Daemon / background process — direct SQLite, no server
- Server-side state or accounts — everything is local
- Multiplayer / trading
- Model-based pet assignment
- Session/CLI wrapper based XP — git commits are the XP source
- Config file — sensible defaults baked in (except optional API key for chat)
