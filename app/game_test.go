package app

import (
	"testing"
	"time"
)

func TestHandleCorrectAndWrong(t *testing.T) {
	s := NewSession("hi")
	s.Handle('h')
	s.Handle('x')
	if s.Pos != 2 {
		t.Fatalf("pos %d", s.Pos)
	}
	if s.Correct() != 1 || s.Attempts() != 2 {
		t.Fatalf("correct %d attempts %d", s.Correct(), s.Attempts())
	}
	if !s.Missed[1] {
		t.Fatal("expected miss on second rune")
	}
	if !s.Done() {
		t.Fatal("should be done")
	}
}

func TestBackspaceKeepsMiss(t *testing.T) {
	s := NewSession("ab")
	s.Handle('x')
	s.Backspace()
	if s.Pos != 0 {
		t.Fatalf("pos %d", s.Pos)
	}
	if !s.Missed[0] {
		t.Fatal("miss should stick")
	}
	s.Handle('a')
	if s.Typed[0] != 'a' {
		t.Fatalf("typed %q", s.Typed[0])
	}
	if s.Attempts() != 2 || s.Correct() != 1 {
		t.Fatalf("attempts %d correct %d", s.Attempts(), s.Correct())
	}
}

func TestClockStartsOnFirstKey(t *testing.T) {
	s := NewSession("aa")
	if s.Started() {
		t.Fatal("clock should wait")
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return now }
	s.Handle('a')
	if !s.Started() {
		t.Fatal("clock should start")
	}
	now = now.Add(time.Minute)
	s.Now = func() time.Time { return now }
	s.Handle('a')
	st := s.Stats()
	// 2 attempts / 5 chars-per-word / 1 minute = 0.4 WPM
	if st.GrossWPM < 0.39 || st.GrossWPM > 0.41 {
		t.Fatalf("wpm %v", st.GrossWPM)
	}
}

func TestWPMFiveCharsOneMinute(t *testing.T) {
	s := NewSession("hello")
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	current := t0
	s.Now = func() time.Time { return current }
	s.Handle('h')
	current = t0.Add(time.Minute)
	s.Now = func() time.Time { return current }
	for _, r := range []rune("ello") {
		s.Handle(r)
	}
	st := s.Stats()
	if st.GrossWPM < 0.99 || st.GrossWPM > 1.01 {
		t.Fatalf("want ~1 WPM, got %v", st.GrossWPM)
	}
	if st.Accuracy != 1 {
		t.Fatalf("accuracy %v", st.Accuracy)
	}
	if st.NetWPM < 0.99 || st.NetWPM > 1.01 {
		t.Fatalf("net %v", st.NetWPM)
	}
}

func TestUTF8Runes(t *testing.T) {
	s := NewSession("café")
	for _, r := range []rune("café") {
		s.Handle(r)
	}
	if !s.Done() || s.Correct() != 4 {
		t.Fatalf("done=%v correct=%d", s.Done(), s.Correct())
	}
}

func TestKeyAndBigramStats(t *testing.T) {
	s := NewSession("th")
	s.Handle('t')
	s.Handle('x')
	ks := s.KeyStats()
	if ks['t'].Attempts != 1 || ks['t'].Misses != 0 {
		t.Fatalf("t stats %+v", ks['t'])
	}
	if ks['h'].Misses != 1 {
		t.Fatalf("h stats %+v", ks['h'])
	}
	bg := s.BigramStats()["th"]
	if bg.Attempts != 1 || bg.Misses != 1 {
		t.Fatalf("bigram %+v", bg)
	}
}

func TestWeakKeys(t *testing.T) {
	stats := map[rune]Stat{
		'p': {Attempts: 10, Misses: 5},
		'a': {Attempts: 10, Misses: 0},
		's': {Attempts: 2, Misses: 2},
		' ': {Attempts: 20, Misses: 20},
	}
	w := WeakKeys(stats, 5, 0.2)
	if len(w) != 1 || w[0] != 'p' {
		t.Fatalf("weak %v", w)
	}
}

func TestBackspaceUnfinishes(t *testing.T) {
	s := NewSession("a")
	s.Handle('a')
	if !s.Done() {
		t.Fatal("done")
	}
	s.Backspace()
	if s.Done() {
		t.Fatal("should unfinish")
	}
}
