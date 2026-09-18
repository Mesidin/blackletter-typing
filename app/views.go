package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width > 0 && m.height > 0 && (m.width < minWidth || m.height < minHeight) {
		return m.viewTooSmall()
	}
	var body string
	switch m.screen {
	case screenProfiles:
		body = m.viewList("Who is typing?", "Pick a profile or make a new one.")
	case screenCreate:
		body = m.viewCreate()
	case screenMenu:
		title := "Main menu"
		sub := "Welcome, typer."
		if m.active != nil {
			title = m.active.Name
			sub = fmt.Sprintf("Level %d  ·  %d XP  ·  %d tests", m.active.Level, m.active.TotalXP, m.active.TotalTestsCompleted)
		}
		body = m.viewList(title, sub)
	case screenPractice:
		body = m.viewList("Literature", "Classic passages. Finish with good accuracy for bonus XP.")
	case screenDrills:
		body = m.viewList("Drills", "Start with the bumps on F and J.")
	case screenTyping:
		body = m.viewTyping()
	case screenSummary:
		body = m.viewSummary()
	case screenHall:
		body = m.viewHall()
	case screenCheat:
		body = m.viewCheat()
	default:
		body = "…"
	}
	head := m.styles.Wordmark(m.width)
	if m.screen == screenTyping || (m.screen == screenSummary && m.boss != nil) {
		head = ""
	}
	return m.styles.frame(m.width, joinBlocks(head, "", body, "", m.helpLine()))
}

func (m Model) viewTooSmall() string {
	msg := fmt.Sprintf("Please make the window bigger.\nNeed at least %d×%d  (now %d×%d)", minWidth, minHeight, m.width, m.height)
	return m.styles.Box.Render(m.styles.Accent.Render(msg))
}

func (m Model) helpLine() string {
	var s string
	switch m.screen {
	case screenTyping:
		if m.boss != nil {
			s = "type to strike   backspace fix   esc flee"
		} else {
			s = "type the glowing letter   backspace fix   esc finish"
		}
	case screenCreate:
		s = "type your name   enter go   esc back   ctrl+c quit"
	case screenCheat:
		s = "type a level   enter go   esc cancel   ctrl+c quit"
	case screenSummary:
		s = "enter continue   esc menu   ctrl+c quit"
	case screenHall:
		s = "esc menu   ctrl+c quit"
	default:
		s = "↑↓ move   enter choose   esc back   ctrl+c quit"
	}
	return m.styles.Help.Render(s)
}

func (m Model) viewList(title, sub string) string {
	head := m.styles.Accent.Render(title) + "\n" + m.styles.Subtle.Render(sub)
	if m.screen == screenMenu {
		head += "\n"
	}
	status := ""
	if m.status != "" {
		status = m.styles.Error.Render(m.status)
	}
	return joinBlocks(head, m.list.View(), status)
}

func (m Model) viewCheat() string {
	cur := 1
	if m.active != nil {
		cur = m.active.Level
	}
	var b strings.Builder
	b.WriteString(m.styles.Accent.Render("Skip to level"))
	b.WriteByte('\n')
	b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Now at %d. Type 1–%d (bosses: 10, 20, … 100).", cur, maxLevel)))
	b.WriteString("\n\n")
	b.WriteString(m.cheatInput.View())
	if m.cheatErr != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Error.Render(m.cheatErr))
	}
	return b.String()
}

func (m Model) viewCreate() string {
	var b strings.Builder
	b.WriteString(m.styles.Accent.Render("New typer"))
	b.WriteByte('\n')
	b.WriteString(m.styles.Subtle.Render("Letters and spaces, 2–16 characters."))
	b.WriteString("\n\n")
	b.WriteString(m.nameInput.View())
	if m.createErr != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.Error.Render(m.createErr))
	}
	return b.String()
}

func (m Model) viewTyping() string {
	if m.session == nil {
		return "no session"
	}
	if m.boss != nil {
		return m.viewBossFight()
	}
	st := m.session.Stats()
	inner := m.width - 8
	if inner < 20 {
		inner = 20
	}

	var b strings.Builder
	b.WriteString(m.styles.Accent.Render(m.sessionTitle))
	if m.sessionHint != "" {
		b.WriteByte('\n')
		b.WriteString(m.styles.Subtle.Render(m.sessionHint))
	}
	b.WriteString("\n\n")
	b.WriteString(m.renderPassage(inner))
	b.WriteString("\n\n")

	wpm := fmt.Sprintf("%.0f", st.GrossWPM)
	acc := fmt.Sprintf("%.0f%%", st.Accuracy*100)
	if !m.session.Started() {
		wpm, acc = "—", "—"
	}
	stats := fmt.Sprintf("WPM %s    accuracy %s    %d/%d", wpm, acc, m.session.Pos, len(m.session.Target))
	b.WriteString(m.styles.Value.Render(stats))

	next := m.session.NextRune()
	if next != 0 {
		label := visibleRune(next)
		finger := FingerFor(next)
		hint := "next: " + label
		if finger != "" {
			hint += "  ·  " + finger
		}
		b.WriteString("\n")
		b.WriteString(m.styles.Accent.Render(hint))
	}

	if m.showKeyboard {
		b.WriteString("\n\n")
		b.WriteString(m.renderKeyboard(next))
	}
	if !m.session.Started() {
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render("Start typing whenever you're ready."))
	}
	return b.String()
}

func (m Model) viewBossFight() string {
	st := m.session.Stats()
	inner := m.width - 8
	if inner < 20 {
		inner = 20
	}
	f := m.boss
	pct := 0.0
	if f.MaxHP > 0 {
		pct = float64(f.HP) / float64(f.MaxHP)
	}
	clock, wpm, acc := "—", "—", "—"
	if m.session.Started() {
		clock = FormatClock(f.Remaining(m.session.Elapsed()))
		wpm = fmt.Sprintf("%.0f", st.GrossWPM)
		acc = fmt.Sprintf("%.0f%%", st.Accuracy*100)
	}
	combo := ""
	if f.Combo >= 2 {
		combo = fmt.Sprintf("\ncombo x%d", f.Combo)
	}
	artW, _ := blockSize(strings.Trim(f.Spec.Art, "\n"))
	barW := 24
	if inner >= artW+24 {
		leftW := inner - artW - 2
		if leftW-2 < barW {
			barW = leftW - 2
		}
	}
	if barW < 8 {
		barW = 8
	}
	info := joinBlocks(
		m.styles.Subtle.Render(f.Spec.Source),
		m.styles.Error.Render(fmt.Sprintf("HP  %d / %d", f.HP, f.MaxHP)),
		hpMeter(pct, barW, m.styles.Error, m.styles.Muted),
		m.styles.Label.Render(fmt.Sprintf("need  %.0f WPM  ·  %.0f%%", f.Spec.MinWPM, f.Spec.MinAccuracy*100)),
		m.styles.Value.Render(fmt.Sprintf("time %s\nWPM %s    acc %s%s", clock, wpm, acc, combo)),
	)
	// Banner is 6 rows; leave room under the portrait for the passage.
	maxArt := m.height - 18
	if maxArt < 8 {
		maxArt = 8
	}
	header := m.bossHeader(f.Spec, info, inner, maxArt, false)
	passageW := inner
	if w := lipgloss.Width(header); w > 0 && w < inner {
		passageW = w
	}
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(m.renderPassage(passageW))
	if !m.session.Started() {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Subtle.Render("Type to strike. Misses deal no damage."))
	}
	return b.String()
}

// bossHeader draws the huge name across the top, then stats on the left
// and the portrait on the right.
func (m Model) bossHeader(spec BossSpec, leftExtra string, width, maxArtH int, muted bool) string {
	title := m.bossNameBlock(spec, width)
	art := strings.Trim(spec.Art, "\n")
	art = clipLines(art, maxArtH)
	artStyle := m.styles.Accent
	if muted {
		artStyle = m.styles.Muted
	}
	artW, _ := blockSize(art)
	gap := 2
	minLeft := 22
	if art != "" && width >= artW+minLeft+gap {
		leftW := width - artW - gap
		leftCol := lipgloss.NewStyle().Width(leftW).MaxWidth(leftW).MarginRight(gap).Render(leftExtra)
		right := artStyle.Render(art)
		return joinBlocks(title, lipgloss.JoinHorizontal(lipgloss.Top, leftCol, right))
	}
	return joinBlocks(title, leftExtra)
}

func (m Model) bossNameBlock(spec BossSpec, width int) string {
	if spec.Banner != "" && bannerFits(spec.Banner, width) {
		return m.styles.Title.Render(spec.Banner)
	}
	return m.styles.Title.Render(spec.Title)
}

func blockSize(s string) (w, h int) {
	if s == "" {
		return 0, 0
	}
	lines := strings.Split(s, "\n")
	h = len(lines)
	for _, line := range lines {
		n := lipgloss.Width(line)
		if n > w {
			w = n
		}
	}
	return w, h
}

func hpMeter(pct float64, width int, fill, empty lipgloss.Style) string {
	if width < 6 {
		width = 6
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	n := int(pct*float64(width) + 0.5)
	if n > width {
		n = width
	}
	return fill.Render(strings.Repeat("█", n)) + empty.Render(strings.Repeat("░", width-n))
}

func clipLines(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= max {
		return s
	}
	return strings.Join(lines[:max], "\n")
}

func (m Model) renderPassage(width int) string {
	s := m.session
	var b strings.Builder
	col := 0
	for i, r := range s.Target {
		if col >= width {
			b.WriteByte('\n')
			col = 0
		}
		if r == '\n' {
			b.WriteByte('\n')
			col = 0
			continue
		}
		glyph := string(r)
		if r == ' ' {
			if i == s.Pos {
				glyph = "_"
			}
		}
		styled := m.styles.Muted.Render(glyph)
		switch {
		case i == s.Pos:
			styled = m.styles.Caret.Render(glyph)
		case i < s.Pos:
			typed := rune(0)
			if i < len(s.Typed) {
				typed = s.Typed[i]
			}
			if typed == r {
				styled = m.styles.Correct.Render(string(r))
			} else {
				show := r
				if typed != 0 {
					show = typed
				}
				styled = m.styles.Wrong.Render(string(show))
			}
		}
		b.WriteString(styled)
		col++
	}
	return b.String()
}

func (m Model) renderKeyboard(next rune) string {
	return RenderKeyboard(m.styles, next, m.width-4, m.height)
}

func (m Model) renderBossTitle(spec BossSpec) string {
	inner := m.width - 6
	if inner < 20 {
		inner = 20
	}
	return m.bossNameBlock(spec, inner)
}

func (m Model) viewSummary() string {
	if m.session == nil || m.active == nil {
		return "done"
	}
	st := m.session.Stats()
	var b strings.Builder
	if m.lastBossFail {
		b.WriteString(m.styles.Error.Render(m.praise))
	} else {
		b.WriteString(m.styles.Success.Render(m.praise))
	}
	b.WriteString("\n\n")
	if m.boss != nil {
		inner := m.width - 8
		if inner < 20 {
			inner = 20
		}
		extra := m.styles.Subtle.Render(m.boss.Spec.Source)
		b.WriteString(m.bossHeader(m.boss.Spec, extra, inner, m.height-16, m.lastBossFail))
	} else {
		b.WriteString(m.styles.Accent.Render(m.sessionTitle))
	}
	b.WriteString("\n\n")
	b.WriteString(row(m.styles, "Net WPM", fmt.Sprintf("%.1f", st.NetWPM)))
	b.WriteString(row(m.styles, "Accuracy", fmt.Sprintf("%.0f%%", st.Accuracy*100)))
	b.WriteString(row(m.styles, "XP earned", fmt.Sprintf("+%d", m.lastXP)))
	if m.lastLitBonus {
		b.WriteString(m.styles.Success.Render("Literature bonus — double XP!") + "\n")
	}
	if m.lastBossWin {
		b.WriteString(m.styles.Success.Render("Boss down! Extra XP.") + "\n")
	}
	if m.lastBossFail {
		b.WriteString(m.styles.Subtle.Render("No XP this time. Continue to fight again.") + "\n")
	}
	if m.finishedEarly && !m.lastBossFail {
		b.WriteString(m.styles.Subtle.Render("Finished early — still counts.") + "\n")
	}
	b.WriteByte('\n')
	if m.lastLevelUp {
		b.WriteString(m.styles.Success.Render(fmt.Sprintf("LEVEL UP! You are now level %d", m.active.Level)))
		b.WriteByte('\n')
		if m.unlockNote != "" {
			b.WriteString(m.styles.Accent.Render(m.unlockNote))
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	need := m.active.XPNeeded()
	into := m.active.XPIntoLevel
	pct := 0.0
	if need > 0 {
		pct = float64(into) / float64(need)
	}
	b.WriteString(m.styles.Label.Render(fmt.Sprintf("Level %d  %d / %d XP", m.active.Level, into, need)))
	b.WriteByte('\n')
	b.WriteString(m.xpBar.ViewAs(pct))

	weak := WeakKeys(m.session.KeyStats(), 3, 0.2)
	if len(weak) > 0 {
		keys := make([]string, 0, len(weak))
		for _, r := range weak {
			keys = append(keys, visibleRune(r))
		}
		b.WriteString("\n\n")
		b.WriteString(m.styles.Subtle.Render("Tricky keys this round: " + strings.Join(keys, "  ")))
	}
	return b.String()
}

func (m Model) viewHall() string {
	rows := m.store.HallOfFame()
	var b strings.Builder
	b.WriteString(m.styles.Accent.Render("Hall of Fame"))
	b.WriteByte('\n')
	b.WriteString(m.styles.Subtle.Render("Local legends only."))
	b.WriteString("\n\n")
	if len(rows) == 0 {
		b.WriteString(m.styles.Muted.Render("No profiles yet."))
		return b.String()
	}
	header := fmt.Sprintf("%-16s %5s %7s %8s %6s", "NAME", "LVL", "XP", "BEST WPM", "TESTS")
	b.WriteString(m.styles.Label.Render(header))
	b.WriteByte('\n')
	for _, p := range rows {
		best := 0.0
		for _, sc := range p.HighScores {
			if sc.WPM > best {
				best = sc.WPM
			}
		}
		mark := "  "
		if m.active != nil && p.ID == m.active.ID {
			mark = "> "
		}
		line := fmt.Sprintf("%s%-14s %5d %7d %8.1f %6d", mark, clip(p.Name, 14), p.Level, p.TotalXP, best, p.TotalTestsCompleted)
		if mark == "> " {
			b.WriteString(m.styles.Accent.Render(line))
		} else {
			b.WriteString(m.styles.Value.Render(line))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func row(st Styles, label, value string) string {
	return st.Label.Render(fmt.Sprintf("%-12s", label)) + "  " + st.Value.Render(value) + "\n"
}

func visibleRune(r rune) string {
	if r == ' ' {
		return "space"
	}
	return string(r)
}

func clip(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n-1]) + "…"
}
