package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLevelForTotalXP(t *testing.T) {
	cases := []struct {
		total, level, into, needed int
	}{
		{0, 1, 0, 80},
		{79, 1, 79, 80},
		{80, 2, 0, 160},
		{240, 3, 0, 240},
		{241, 3, 1, 240},
	}
	for _, c := range cases {
		level, into, needed := LevelForTotalXP(c.total)
		if level != c.level || into != c.into || needed != c.needed {
			t.Fatalf("xp %d: got %d/%d/%d want %d/%d/%d",
				c.total, level, into, needed, c.level, c.into, c.needed)
		}
	}
}

func TestMaxUnlockedTier(t *testing.T) {
	if MaxUnlockedTier(1) != 0 || MaxUnlockedTier(2) != 1 || MaxUnlockedTier(3) != 2 || MaxUnlockedTier(5) != 3 {
		t.Fatalf("unlocks 1=%d 2=%d 3=%d 5=%d", MaxUnlockedTier(1), MaxUnlockedTier(2), MaxUnlockedTier(3), MaxUnlockedTier(5))
	}
}

func TestComputeXP(t *testing.T) {
	st := SessionStats{Attempts: 40, Correct: 40, Accuracy: 1, GrossWPM: 20}
	xp := ComputeXP(st, true, false)
	// 40 * 1 * (1+20/80) * 1.15 ≈ 57.5; float rounding may land on 57 or 58
	if xp < 57 || xp > 58 {
		t.Fatalf("perfect xp %d", xp)
	}
	early := ComputeXP(st, false, false)
	if early >= xp {
		t.Fatalf("early %d should be less than finished %d", early, xp)
	}
	if ComputeXP(SessionStats{}, true, false) != 0 {
		t.Fatal("empty session should award 0")
	}
	min := ComputeXP(SessionStats{Attempts: 1, Correct: 0, Accuracy: 0, GrossWPM: 0}, false, false)
	if min != 1 {
		t.Fatalf("minimum xp %d", min)
	}
	lit := ComputeXP(st, true, true)
	if lit < xp*2-1 || lit > xp*2+1 {
		t.Fatalf("literature bonus %d vs base %d", lit, xp)
	}
	noBonus := ComputeXP(SessionStats{Attempts: 10, Correct: 5, Accuracy: 0.5, GrossWPM: 10}, true, true)
	plain := ComputeXP(SessionStats{Attempts: 10, Correct: 5, Accuracy: 0.5, GrossWPM: 10}, true, false)
	if noBonus != plain {
		t.Fatalf("low accuracy literature should not get bonus: %d vs %d", noBonus, plain)
	}
}

func TestBetterScore(t *testing.T) {
	a := Score{XP: 10, Accuracy: 0.9, WPM: 10}
	b := Score{XP: 9, Accuracy: 1, WPM: 40}
	if !BetterScore(a, b) {
		t.Fatal("xp should win")
	}
}

func TestValidateName(t *testing.T) {
	if ValidateName("A") == nil {
		t.Fatal("too short")
	}
	if ValidateName("Ada") != nil {
		t.Fatal("Ada should pass")
	}
	if ValidateName("Ada 2") == nil {
		t.Fatal("digits rejected")
	}
	if ValidateName(" ann ") != nil {
		t.Fatal("trimmed letters should pass")
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BLACKLETTER_HOME", dir)
	s, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Create("Ada")
	if err != nil {
		t.Fatal(err)
	}
	s.Award(p.ID, "drill:home-fj", 50, Score{WPM: 12, Accuracy: 0.9, XP: 50},
		map[rune]Stat{'f': {Attempts: 10, Misses: 3}},
		map[string]Stat{"fj": {Attempts: 4, Misses: 2}},
	)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	s2, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	p2 := s2.Profile(p.ID)
	if p2 == nil || p2.Name != "Ada" {
		t.Fatal("missing profile")
	}
	if p2.TotalXP != 50 || p2.Level != 1 {
		t.Fatalf("xp/level %+v", p2)
	}
	if p2.HighScores["drill:home-fj"].XP != 50 {
		t.Fatalf("high score %+v", p2.HighScores)
	}
	if p2.KeyStats["f"].Misses != 3 {
		t.Fatalf("key stats %+v", p2.KeyStats)
	}
}

func TestCorruptFileNotClobbered(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("BLACKLETTER_HOME", dir)
	path := filepath.Join(dir, "profiles.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	if s.LoadErr == nil {
		t.Fatal("expected load error")
	}
	if err := s.Save(); err == nil {
		t.Fatal("save should refuse")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "{not json" {
		t.Fatalf("file was clobbered: %q", b)
	}
}

func TestHallOfFameSort(t *testing.T) {
	s := &Store{okToSave: true}
	s.File.Profiles = []Profile{
		{Name: "Leo", Level: 2, TotalXP: 100},
		{Name: "Ada", Level: 4, TotalXP: 500},
		{Name: "Bea", Level: 4, TotalXP: 400},
	}
	hof := s.HallOfFame()
	if hof[0].Name != "Ada" || hof[1].Name != "Bea" || hof[2].Name != "Leo" {
		t.Fatalf("order %+v %+v %+v", hof[0].Name, hof[1].Name, hof[2].Name)
	}
}

func TestHighScoreReplacedOnlyIfBetter(t *testing.T) {
	s := &Store{okToSave: true}
	p, _ := s.Create("Ada")
	s.Award(p.ID, "x", 10, Score{XP: 10, Accuracy: 0.9, WPM: 10}, nil, nil)
	s.Award(p.ID, "x", 5, Score{XP: 5, Accuracy: 1, WPM: 40}, nil, nil)
	got := s.Profile(p.ID).HighScores["x"]
	if got.XP != 10 {
		t.Fatalf("kept worse score %+v", got)
	}
	s.Award(p.ID, "x", 20, Score{XP: 20, Accuracy: 0.8, WPM: 8}, nil, nil)
	got = s.Profile(p.ID).HighScores["x"]
	if got.XP != 20 {
		t.Fatalf("did not replace %+v", got)
	}
}
