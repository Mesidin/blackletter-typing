package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyRunes(s string) []tea.Msg {
	var out []tea.Msg
	for _, r := range s {
		out = append(out, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return out
}

func apply(t *testing.T, m Model, msgs ...tea.Msg) Model {
	t.Helper()
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		var ok bool
		m, ok = next.(Model)
		if !ok {
			t.Fatalf("model type %T", next)
		}
	}
	return m
}

func testModel(t *testing.T) Model {
	t.Helper()
	t.Setenv("BLACKLETTER_HOME", t.TempDir())
	store, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	passages := []Passage{
		{ID: "t1", Title: "Tiny", Source: "test", Category: "literature", Tier: 1, Text: "hi"},
	}
	m := New(store, passages, NewStyles(CyberpunkPalette()))
	m.width, m.height = 80, 24
	return m
}

func TestCreateProfileThenDrillSession(t *testing.T) {
	m := testModel(t)
	if m.screen != screenCreate {
		t.Fatalf("screen %d", m.screen)
	}
	m = apply(t, m, keyRunes("Ada")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenMenu {
		t.Fatalf("after create screen %d", m.screen)
	}
	if m.active == nil || m.active.Name != "Ada" {
		t.Fatalf("active %+v", m.active)
	}

	m = apply(t, m, tea.KeyMsg{Type: tea.KeyDown}) // Drills
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenDrills {
		t.Fatalf("drills screen %d", m.screen)
	}

	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // first letter drill
	if m.screen != screenTyping || m.session == nil {
		t.Fatalf("typing screen %d session %v", m.screen, m.session)
	}

	for _, r := range append([]rune(nil), m.session.Target...) {
		m = apply(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if m.screen != screenSummary {
		t.Fatalf("summary screen %d pos %d/%d", m.screen, m.session.Pos, len(m.session.Target))
	}
	if m.lastXP < 1 {
		t.Fatalf("xp %d", m.lastXP)
	}
	if m.active.TotalTestsCompleted != 1 {
		t.Fatalf("tests %d", m.active.TotalTestsCompleted)
	}

	view := m.View()
	if view == "" {
		t.Fatal("empty view")
	}
}

func TestPracticeLockedAtLevelOne(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Bea")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	// Play, Drills, Literature
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenMenu {
		t.Fatalf("should stay on menu, got %d", m.screen)
	}
	if m.status == "" {
		t.Fatal("expected locked message")
	}
}

func TestEscAbandonsEmptySession(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Cal")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}) // drills
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyDown}, tea.KeyMsg{Type: tea.KeyEnter}) // first curriculum drill
	if m.screen != screenTyping {
		t.Fatalf("screen %d", m.screen)
	}
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != screenMenu {
		t.Fatalf("empty esc should skip summary, got %d", m.screen)
	}
}

func TestPlayStartsFirstLetterDrill(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Dot")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // Play
	if m.screen != screenTyping {
		t.Fatalf("screen %d", m.screen)
	}
	if m.sessionKind != kindPlay {
		t.Fatalf("kind %d", m.sessionKind)
	}
	if m.journeyStepID == "" || !strings.Contains(m.sessionTitle, "Index") {
		t.Fatalf("title %q step %q", m.sessionTitle, m.journeyStepID)
	}
	for _, r := range append([]rune(nil), m.session.Target...) {
		m = apply(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if m.screen != screenSummary {
		t.Fatalf("summary %d", m.screen)
	}
	if m.active.JourneyCount(m.journeyStepID) != 1 {
		t.Fatalf("journey count %d", m.active.JourneyCount(m.journeyStepID))
	}

	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.screen != screenTyping {
		t.Fatalf("continue should stay on Play, got screen %d", m.screen)
	}
	if m.sessionKind != kindPlay {
		t.Fatalf("continue kind %d", m.sessionKind)
	}
}

func TestSummaryEscGoesToMenu(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Eve")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	for _, r := range append([]rune(nil), m.session.Target...) {
		m = apply(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != screenMenu {
		t.Fatalf("esc should menu, got %d", m.screen)
	}
}

func TestPlayAtBossLevelStartsFight(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Fay")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m.active.Level = 10
	m.active.TotalXP = 3600
	m.store.profilePtr(m.active.ID).Level = 10
	m.store.profilePtr(m.active.ID).TotalXP = 3600
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.sessionKind != kindBoss || m.boss == nil {
		t.Fatalf("kind %d boss %v", m.sessionKind, m.boss)
	}
	if m.boss.Spec.Name != "Wolfman" {
		t.Fatalf("boss %q", m.boss.Spec.Name)
	}
	if len(m.session.Target) < 80 {
		t.Fatalf("boss text too short: %d", len(m.session.Target))
	}
}
