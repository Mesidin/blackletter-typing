# Comprehensive Project Specification: Cyberpunk Terminal Typing Tutor (Go + Bubbletea)

## 1. Overview & Vision
Build a standalone, retro-cyberpunk TUI typing tutor in Go using the `github.com/charmbracelet/bubbletea` framework and `github.com/charmbracelet/lipgloss` for styling. The application is tailored for absolute beginner typists (such as young kids) and scales naturally as they grow into advanced typists. 

It runs smoothly across macOS and Linux/Omarchy environments, supports dynamic OS-level theme integration (respecting Omarchy color schemes), tracks multiple user profiles with high scores and leveling systems, incorporates progressive training drills, and pulls text material from classic literature, poetry, and scripture.

---

## 2. Directory Structure
```text
typer/
├── go.mod
├── go.sum
├── main.go
├── app/
│   ├── model.go       # State definitions & Bubbletea lifecycle (Init, Update, View)
│   ├── styles.go      # Dynamic cyberpunk theme definitions (OS-aware)
│   ├── game.go        # Typing evaluation engine & real-time stats tracking
│   ├── profile.go     # Profile management, XP, leveling, and persistent high scores JSON
│   ├── corpus.go      # Text loader & intelligent difficulty-tier categorizer
│   └── drills.go      # Adaptive letter-frequency and hand-placement training drills
└── assets/
    └── texts/
        ├── literature/  # e.g., Alice in Wonderland, Frankenstein (chunked by difficulty)
        ├── poetry/      # e.g., Dickinson, Poe (rhythmic short lines)
        └── scripture/   # e.g., Proverbs, Psalms (short wisdom verses)
```

---

## 3. Technical & Architectural Requirements

### A. Dynamic OS Theme Integration (`styles.go`)
*   **Omarchy/Linux Compatibility:** Check for environment variables or standard terminal color schemes (e.g., parsing Xresources, OMR_THEME env vars, or fallback ANSI sequences). 
*   **Fallback Cyberpunk Aesthetic:** Pitch black background (`#0d0208`), Matrix green (`#00ff66`) for correct text, Neon cyan (`#00f0ff`) for borders/UI framing, Hot pink (`#ff0055`) for errors/warnings, and Dim gray (`#333333`) for un-typed text.

### B. Gamification, Progression, & Profiles (`profile.go`)
*   **Profiles:** Allow kids to select or create a profile on launch (stored in a local `profiles.json` file). Profiles track:
    *   `Name`, `Level`, `CurrentXP`, `TotalTestsCompleted`.
    *   `HighScores`: Map of category/level best scores.
*   **XP & Leveling System:** XP is gained based on WPM, accuracy, and text length. Leveling up unlocks higher difficulty tiers (e.g., Level 1: Micro-chunks & home-row drills; Level 5+: Complex poetry and full paragraphs).
*   **Scoreboard / Comparison:** A dedicated "Hall of Fame" view where kids can compare their high scores, levels, and XP side-by-side.

### C. Progressive Training & Hand Placement (`drills.go`)
*   **Targeted Drills:** As users grow, introduce common typing drills emphasizing tricky letter combinations, home-row mastery (`asdf jkl;`), and common word structures derived from frequency analysis of the corpus text.

### D. Content Pipeline & Recommendations (`corpus.go`)
*   **Difficulty Tiers:**
    *   *Tier 1 (Beginner):* Short words, home-row focus, sentences under 30 characters.
    *   *Tier 2 (Intermediate):* Standard sentence structure from classic literature/scripture.
    *   *Tier 3 (Advanced):* Complex punctuation, multi-clause poetry, and long technical/literary passages.
*   The app recommends texts based on the user's current level and past performance metrics.

---

## 4. Implementation Instructions for AI Build Tool (Grok)

Feed these phases sequentially to Grok:

1. **Phase 1:** Initialize Go module (`go mod init typer`), install dependencies (`go get github.com/charmbracelet/bubbletea github.com/charmbracelet/lipgloss`), and create folder scaffolding.
2. **Phase 2:** Implement `app/styles.go` with dynamic color parsing and cyberpunk borders.
3. **Phase 3:** Implement `app/profile.go` for persistent JSON profiles, XP calculation, level progression, and scoreboard sorting.
4. **Phase 4:** Implement `app/corpus.go` and `app/drills.go` to handle text loading, filtering by difficulty, and generating hand-placement drills.
5. **Phase 5:** Implement `app/game.go` and `app/model.go` managing the Bubbletea state machine (Profile Select -> Main Menu -> Category/Drill Select -> Typing Session -> Summary/XP Screen -> Scoreboard).
6. **Phase 6:** Wire everything in `main.go` and create initial sample JSON datasets under `assets/texts/`.


---

## 5. Differentiation from `Typearchy` (`tsouth89/typearchy`)
While drawing inspiration from `Typearchy` [cite: 1.1.1] for its local-first architecture, clean Omarchy theme integration, and bar launcher/widget design, this application pivots heavily toward **pedagogical teaching and kids' gamification** rather than developer code practice [cite: 1.1.1] and raw speedruns:
*   **Target Audience:** Absolute beginner children (focusing on finger placement, home-row mastery, and progressive unlocking) rather than developers practicing Bash/Python/Rust code [cite: 1.1.1].
*   **Adaptive Error Tracking:** Instead of just logging speed stats, the engine actively identifies weak keys/bigrams and auto-generates custom remediation mini-drills.
*   **Kid-Friendly Gamification:** Experience points (XP), character leveling, unlocks for new classic literature tiers, and a local side-by-side profile Hall of Fame.
