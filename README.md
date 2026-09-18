# Blackletter

A gothic typing tutor for the terminal. Kids learn the home row, then nonsense words, then real sentences. Classic public-domain literature waits in its own mode. Every tenth level is a boss from the old stories — Wolfman, Hyde, Frankenstein’s Monster, Dracula, Ahab’s whale, Grendel, the Raven, the Headless Horseman, the Kraken, Leviathan.

Built in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Local only: no accounts, no network.

```bash
go run .
```

## Play

**Play** is a guided path. A step counts when you finish it at 80% accuracy or better.

| Stage | What you type |
|---|---|
| Letter drills | Home-row keys, in order (`F`/`J` first). Level 1 asks for each drill twice. |
| Nonsense words | Made-up 3–5 letter words from keys you know |
| Short sentences | Real sentences, from level 2 |

When the path for a level is done, Play keeps offering bonus rounds.

**Literature** is optional: Alice, Frankenstein, Dickinson, Poe, Psalms, Proverbs. Finish a passage at 80%+ for double XP.

After a lesson, **enter** continues the same track. **esc** returns to the menu.

## Bosses

Levels 10, 20, … 100 are boss fights: a longer review, a clock, and a portrait. Correct keys deal damage. Misses deal none. You must also hit a minimum WPM and accuracy. Fail or run out of time and the boss still stands — Continue retries. Win for extra XP.

## Levels

XP to finish a level is `80 × current level`, so later levels take more typing.

| Level | Unlocks |
|---|---|
| 1 | Letter drills and nonsense words |
| 2 | Short sentences in Play, plus Literature |
| 3 | Longer literature and scripture |
| 5 | Poetry and long passages |
| 10, 20, … 100 | Boss fights |

Cap is level 100.

## Keys

| Key | Action |
|---|---|
| `↑` `↓` | Move in menus |
| `enter` | Choose / continue on the same track |
| `esc` | Back to menu (during a lesson: finish early) |
| `backspace` | Fix a letter |
| `ctrl+c` | Quit |

Font size is the terminal’s (`Cmd++` / `Cmd+-` on a Mac). The giant titles are drawn with box-drawing characters, not a special font.

## Install on another machine

Go compiles to **one file**. The other computer does not need Go.

```bash
make dist
```

| File | For |
|---|---|
| `dist/blackletter-mac-apple` | Mac with Apple Silicon (M1–M4) |
| `dist/blackletter-mac-intel` | Intel Mac |
| `dist/blackletter-linux` | Omarchy / most Linux PCs |

**Mac (Command Line & Dock)**

```bash
chmod +x blackletter-mac-apple
xattr -d com.apple.quarantine blackletter-mac-apple   # if macOS blocks it
./blackletter-mac-apple
```

To install as a macOS `.app` bundle for your Dock (matching `chess-with-go` & `go-with-go`):

```bash
make app    # or: sh scripts/macos-app.sh
cp -R "Blackletter.app" /Applications/
```

Then open `/Applications` in Finder and drag `Blackletter` onto your Mac Dock.

**Omarchy / Linux (CLI & Apps Menu)**

```bash
chmod +x blackletter-linux
sudo mv blackletter-linux /usr/local/bin/blackletter
blackletter
```

To add Blackletter to your **Omarchy / Linux main apps menu**:

```bash
make install-desktop    # or: sh scripts/install-desktop.sh
```

This installs the binary to `~/.local/bin/blackletter` and creates `~/.local/share/applications/blackletter.desktop` with `Terminal=true` so it automatically launches in a terminal when selected from Omarchy's application launcher.

On Omarchy it reads the current theme from `colors.toml`. USB, AirDrop, `scp`, or a shared folder all work.

Profiles are stored at `~/.blackletter/profiles.json`. Override the directory with `BLACKLETTER_HOME` (tests do this). Optional theme file: `BLACKLETTER_THEME=/path/to/colors.toml`.

## Develop

Needs Go 1.22+ (built with 1.27).

```bash
make test
make run
```

See [AGENTS.md](AGENTS.md) if you are an agent (or a human) picking the project up later.
