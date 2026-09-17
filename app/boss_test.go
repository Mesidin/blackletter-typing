package app

import (
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func TestIsBossLevel(t *testing.T) {
	for _, lv := range []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100} {
		if !IsBossLevel(lv) {
			t.Fatalf("%d should be a boss level", lv)
		}
	}
	for _, lv := range []int{0, 1, 9, 11, 15, 99, 110} {
		if IsBossLevel(lv) {
			t.Fatalf("%d should not be a boss level", lv)
		}
	}
}

func TestBossRosterLiterary(t *testing.T) {
	r := BossRoster()
	if len(r) != 10 {
		t.Fatalf("got %d bosses", len(r))
	}
	want := []string{"Wolfman", "Hyde", "the Monster", "Dracula", "the Whale"}
	for i, name := range want {
		if r[i].Name != name {
			t.Fatalf("boss %d name %q want %q", r[i].Level, r[i].Name, name)
		}
		if r[i].Art == "" {
			t.Fatalf("%s missing art", name)
		}
		if r[i].MinWPM < 10 || r[i].MinAccuracy < 0.8 {
			t.Fatalf("%s bars too low: %+v", name, r[i])
		}
	}
	if BossFor(100).Name != "Leviathan" {
		t.Fatalf("final boss %q", BossFor(100).Name)
	}
	for _, b := range r {
		if b.Banner == "" {
			t.Fatalf("%s missing name banner", b.Name)
		}
	}
	kraken, levi := BossFor(90).Art, BossFor(100).Art
	if kraken == levi {
		t.Fatal("kraken and leviathan art should not match")
	}
	if !strings.Contains(kraken, "⣿") || !strings.Contains(levi, "⣿") {
		t.Fatal("boss art should use dense braille blocks")
	}
	later := BossFor(20).MinWPM
	if later <= BossFor(10).MinWPM {
		t.Fatal("later bosses should demand more WPM")
	}
}

func TestHitDamage(t *testing.T) {
	if HitDamage(false, 10, 80, 15) != 0 {
		t.Fatal("errors deal no damage")
	}
	if HitDamage(true, 0, 0, 15) != 1 {
		t.Fatal("base correct hit is 1")
	}
	if HitDamage(true, 4, 0, 15) != 2 {
		t.Fatal("combo should add damage")
	}
	if HitDamage(true, 4, 20, 15) != 3 {
		t.Fatal("combo + WPM should add damage")
	}
}

func TestBossEvaluate(t *testing.T) {
	spec := BossSpec{Name: "Hyde", Title: "Mr. Hyde", MinWPM: 20, MinAccuracy: 0.9}
	f := &BossFight{Spec: spec, HP: 0, MaxHP: 10, Limit: time.Minute}
	st := SessionStats{GrossWPM: 25, Accuracy: 0.95, Done: true}
	if win, _ := f.Evaluate(st, true, time.Second); !win {
		t.Fatal("should win")
	}
	if win, reason := f.Evaluate(st, true, 2*time.Minute); win || reason != "time" {
		t.Fatalf("timeout: win=%v reason=%s", win, reason)
	}
	slow := st
	slow.GrossWPM = 10
	if win, reason := f.Evaluate(slow, true, time.Second); win || reason != "slow" {
		t.Fatalf("slow: win=%v reason=%s", win, reason)
	}
	miss := st
	miss.Accuracy = 0.5
	if win, reason := f.Evaluate(miss, true, time.Second); win || reason != "errors" {
		t.Fatalf("errors: win=%v reason=%s", win, reason)
	}
	f.HP = 3
	if win, reason := f.Evaluate(st, true, time.Second); win || reason != "alive" {
		t.Fatalf("alive: win=%v reason=%s", win, reason)
	}
}

func TestBuildBossTextIsLongReview(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	ps := []Passage{
		{ID: "s", Kind: "sentence", Tier: 1, Text: "The cat sat on the mat."},
		{ID: "l", Kind: "literature", Tier: 2, Text: "Call me Ishmael."},
	}
	text := BuildBossText(10, ps, rng)
	if len([]rune(text)) < 100 {
		t.Fatalf("too short: %d", len([]rune(text)))
	}
	if !strings.Contains(text, "cat") && !strings.Contains(text, " ") {
		t.Fatalf("expected review words: %q", text)
	}
}

func TestBossTimeLimit(t *testing.T) {
	d := BossTimeLimit(100, 20)
	if d < 20*time.Second {
		t.Fatalf("limit %s", d)
	}
}

func TestHasBeatBoss(t *testing.T) {
	p := Profile{BossesBeaten: []int{10, 20}}
	if !p.HasBeatBoss(10) || p.HasBeatBoss(30) {
		t.Fatal("beat map")
	}
}
