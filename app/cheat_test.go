package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFeedCheatSeq(t *testing.T) {
	var buf []string
	var hit bool
	for _, k := range cheatSeq {
		buf, hit = feedCheat(buf, k)
	}
	if !hit {
		t.Fatal("expected hit")
	}
}

func TestFeedCheatIgnoresNoiseThenMatches(t *testing.T) {
	buf, hit := feedCheat(nil, "enter")
	if hit {
		t.Fatal("enter should not cheat")
	}
	for _, k := range cheatSeq {
		buf, hit = feedCheat(buf, k)
	}
	if !hit {
		t.Fatal("expected hit after noise")
	}
}

func TestParseCheatLevel(t *testing.T) {
	n, err := parseCheatLevel("10")
	if err != nil || n != 10 {
		t.Fatalf("got %d %v", n, err)
	}
	if _, err := parseCheatLevel("0"); err == nil {
		t.Fatal("0 should fail")
	}
	if _, err := parseCheatLevel("101"); err == nil {
		t.Fatal("101 should fail")
	}
	if _, err := parseCheatLevel("nope"); err == nil {
		t.Fatal("text should fail")
	}
}

func TestSkipToLevel(t *testing.T) {
	t.Setenv("BLACKLETTER_HOME", t.TempDir())
	s, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Create("Ada")
	if err != nil {
		t.Fatal(err)
	}
	got := s.SkipToLevel(p.ID, 10)
	if got.Level != 10 {
		t.Fatalf("level %d", got.Level)
	}
	want := 0
	for l := 1; l < 10; l++ {
		want += 80 * l
	}
	if got.TotalXP != want {
		t.Fatalf("xp %d want %d", got.TotalXP, want)
	}
}

func TestKonamiOpensPromptThenSkips(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Ada")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	keys := []tea.KeyMsg{
		{Type: tea.KeyUp}, {Type: tea.KeyUp},
		{Type: tea.KeyDown}, {Type: tea.KeyDown},
		{Type: tea.KeyLeft}, {Type: tea.KeyRight},
		{Type: tea.KeyLeft}, {Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune("b")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
	}
	for _, k := range keys {
		m = apply(t, m, k)
	}
	if m.screen != screenCheat {
		t.Fatalf("screen %d", m.screen)
	}
	m = apply(t, m, keyRunes("10")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.active.Level != 10 {
		t.Fatalf("level %d status %q err %q", m.active.Level, m.status, m.cheatErr)
	}
	if m.screen != screenMenu {
		t.Fatalf("screen %d", m.screen)
	}
}

func TestCheatPromptEscCancels(t *testing.T) {
	m := testModel(t)
	m = apply(t, m, keyRunes("Bea")...)
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	keys := []tea.KeyMsg{
		{Type: tea.KeyUp}, {Type: tea.KeyUp},
		{Type: tea.KeyDown}, {Type: tea.KeyDown},
		{Type: tea.KeyLeft}, {Type: tea.KeyRight},
		{Type: tea.KeyLeft}, {Type: tea.KeyRight},
		{Type: tea.KeyRunes, Runes: []rune("b")},
		{Type: tea.KeyRunes, Runes: []rune("a")},
	}
	for _, k := range keys {
		m = apply(t, m, k)
	}
	m = apply(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.screen != screenMenu {
		t.Fatalf("screen %d", m.screen)
	}
	if m.active.Level != 1 {
		t.Fatalf("level should stay 1, got %d", m.active.Level)
	}
}
