# Blackletter — agent notes

Gothic TUI typing tutor for kids. Go + Bubble Tea + Lip Gloss. Local only.

## Commands

```bash
go test ./...          # or: make test
go run .               # or: make run
./scripts/dist.sh      # or: make dist
./scripts/macos-app.sh # or: make app (builds macOS Blackletter.app bundle for Dock)
./scripts/install-desktop.sh # or: make install-desktop (installs .desktop entry for Omarchy apps menu)
gofmt -w .
```

Module path is `blackletter`. Binary name is `blackletter`. Data dir is `~/.blackletter` (`BLACKLETTER_HOME` in tests).

## Layout

| Path | Role |
|---|---|
| `main.go` | Embeds `assets/texts`, opens the store, runs Bubble Tea |
| `app/model.go` | Screen state machine |
| `app/views.go` | Rendering, including boss header (title left, portrait right) |
| `app/game.go` | Typing engine (runes, WPM, accuracy, miss stats) |
| `app/profile.go` | JSON save/load, XP, levels |
| `app/journey.go` | Play path: letter drills → nonsense → sentences |
| `app/drills.go` | Home-row curriculum + generators |
| `app/corpus.go` | Passages (`kind: sentence` vs literature) |
| `app/boss.go` / `boss_art.go` | Decade bosses, braille portraits, fight rules |
| `app/theme.go` / `styles.go` | Omarchy `colors.toml` or ink-and-gold fallback |
| `app/banner.go` | 6-line block-letter banners (box-drawing, not a font) |
| `app/fraktur.go` | Tiny Unicode fraktur (do not use for titles — unreadable) |
| `app/cheat.go` | Konami prompt to skip to a level (dev/testing) |
| `assets/texts/` | Public-domain JSON excerpts |
| `scripts/dist.sh` | Cross-compile Mac Intel, Mac Apple, Linux |

## Rules of the product

- Kids-first copy. Never “failed.” Boss loss is “still stands — try again.”
- **enter** on a summary continues the **same track** (Play, same drill, next literature, retry boss). **esc** goes to the menu.
- Play is a journey, not a random recommended pick. Letter drills first, then nonsense words, then sentences once unlocked.
- Literature is a separate mode with bonus XP for a clean finish (80%+, completed).
- Levels 10, 20, … 100 are bosses. Play is the fight until that boss is beaten.
- Domain packages (`game`, `profile`, `journey`, `drills`, `corpus`, `boss` logic) stay unit-testable without a TTY.

## Adding content

**Sentence** (Play path): JSON in `assets/texts/sentences/` with `"kind": "sentence"` and a `tier` (1–3). Keep them short.

**Literature**: JSON under `literature/`, `poetry/`, or `scripture/` without `kind` (defaults to literature) or `"kind": "literature"`.

**Drill**: append to `Curriculum` in `drills.go`. Play’s letter stage uses `LetterDrills()` (skips `Words: true`).

**Boss**: add to `bossAt` in `boss.go` and a braille portrait in `boss_art.go`. Portraits are converted from high-contrast woodcuts via `scripts/braille.py`. Keep Kraken (tentacles) visually distinct from Leviathan (serpent).

## Theme

Resolution: `NO_COLOR` → `BLACKLETTER_THEME` → Omarchy current `colors.toml` → ink-and-gold fallback. Near-black `muted` is lifted for Mac terminals. Titles use 6-line ANSI Shadow block letters (`BigBanner` in `banner.go`) — drawn with `█╗╔`, not a font and not an image. Do not switch titles to Unicode Fraktur; it is too small to read.

## Cheat (keep out of the player README)

On menus, not during a lesson: `↑ ↑ ↓ ↓ ← → ← → b a` opens a level prompt (1–100). Needs a selected profile.

## Layout gotchas

Boss fights: huge name banner across the top; portrait to the **right** of HP/timer; typing line underneath. Clip the portrait if the window is short. Braille width depends on the font; measure with `lipgloss.Width`.

## Tests

Table-driven tests live next to the code (`*_test.go`). After behavior changes, run `go test ./...`. TUI smoke tests in `model_test.go` drive keys through `Update`.
