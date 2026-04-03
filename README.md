# gitpet

A terminal pet that levels up when you commit code.

```
+======================================+
|  Gobby  -  Goblin  -  Common         |
|                                      |
|      /\  /\                          |
|     ( o  o )                         |
|      > {} <                          |
|     /|/  \|\                         |
|    (_|    |_)                        |
|                                      |
|  LVL 5   ===---------  53/734        |
|                                      |
|  WIT      ==--------  22             |
|  DEPTH    ===-------  36             |
|  STAMINA  ===-------  31             |
|  LUCK     ===-------  37             |
|  ATTUNE   ==--------  23             |
|                                      |
|  Next evo: Lv.30  Stone: no          |
+======================================+
```

## How it works

Every `git commit` feeds your pet. A global post-commit hook calculates XP from your diff — lines changed, files touched, commit size. Small, focused commits on feature branches earn the most XP.

```
Gobby gained 87 XP! (340/544)
Branch bonus: 1.25x
Streak: 5 days (1.3x bonus!)
DROP: Evolution Stone!
```

## Install

```bash
go install github.com/jakebayliss/gitpet/cmd/pet@latest
```

## Quick start

```bash
pet spawn          # hatch your first pet, name it, done
pet show           # see your pet
# now just commit code — XP is automatic
```

That's it. `pet spawn` installs a global git hook. Every commit in any repo feeds your pet.

## Commands

| Command | What it does |
|---------|-------------|
| `pet spawn` | Hatch a new pet (first is free, additional require egg drops) |
| `pet show` | Display your active pet's card |
| `pet watch` | Live-updating pet window (keep it floating while you code) |
| `pet list` | Show all your pets |
| `pet switch <name>` | Change active pet |
| `pet kill <name>` | Permanently delete a pet |
| `pet evolve` | Evolve your pet (requires level + Evolution Stone) |
| `pet prestige` | Reset to level 1 for a prestige star (requires level 99) |
| `pet inv` | Show your inventory |
| `pet log` | Recent commits with XP and drops |
| `pet init` | Reinstall the global git hook |

## XP system

XP is calculated from each commit's diff:

```
base_xp    = lines changed (capped at 200)
file_bonus = files touched * 5 (capped at 10 files)
penalty    = 0.5x if lines changed > 500
```

**Bonuses:**
- **Branch commit** — 1.25x for not committing to main/master
- **Streak** — 1.1x per consecutive day (day 3+), caps at 1.5x

Level cap is **99**. XP curve is RuneScape-inspired — each level costs more than the last.

## Drops

Every commit rolls a drop table:

| Drop | Chance | Effect |
|------|--------|--------|
| XP Crystal | ~1 in 17 | +200 bonus XP |
| Title Scroll | ~1 in 50 | Random title for your pet |
| Evolution Stone | ~1 in 125 | Required to evolve |
| Rare Egg | ~1 in 250 | Hatch a pet from the rare pool |
| Legendary Egg | ~1 in 1000 | Hatch a pet from the legendary pool |

## Species

### Common pool (starter)

| Species | Art |
|---------|-----|
| Axolotl | `( o.o )~` |
| Capybara | `\| ^ ^ \|` |
| Goblin | `( o o )` |
| Pangolin | `/@@@@\___` |
| Tardigrade | `( o o )` |
| Blobfish | `\| . . \|` |
| Mantis Shrimp | `--O O--` |
| Nudibranch | `* * *` |

### Rare pool (~1 in 250 commits)

Salamander, Dire Wolf, Thunderbird, Crystal Fox, Chimera, Shadow Lynx

### Legendary pool (~1 in 1000 commits)

Phoenix, Leviathan, Basilisk, Void Cat, Kraken, Celestial Beetle, Mimic, Wyrm

## Evolution

Pets evolve through 4 stages. Each requires a level threshold **and** an Evolution Stone drop.

| Stage | Level | Stat multiplier |
|-------|-------|-----------------|
| 1 | -- | 1.0x |
| 2 | 30 | 1.3x |
| 3 | 60 | 1.6x |
| 4 | 90 | 2.0x |

Each stage gets bigger ASCII art:

```
Stage 1:          Stage 4:

  /\  /\                /\    /\
 ( o  o )             / OO  OO \
  > {} <             | >>>{}<<<  |
 /|/  \|\            |/  |  |  \|
(_|    |_)          // / |  | \ \\
                   // /  |  |  \ \\
                  (_/ /  |  |\  \_)
                      ~~~~~~~~
                     GOBLIN EMPEROR
                    fear the greenkin
```

## Prestige

Hit level 99 and you can prestige — reset to level 1 in exchange for:

- A permanent star on your pet's name
- +2 to all base stats
- Evolution resets to stage 1

No prestige cap. Grind forever.

## `pet watch`

A live Bubbletea-powered window you keep floating while you work. Polls the database every 2 seconds and reacts to XP changes, level ups, and drops from commits in other terminals.

```bash
pet watch
```

Press `q` to quit.

## Data

Everything is local. Single SQLite database at `~/.petd/pets.db`. No server, no accounts, no network calls.

## Built with

- [Go](https://go.dev)
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [modernc.org/sqlite](https://modernc.org/sqlite) — Pure Go SQLite
