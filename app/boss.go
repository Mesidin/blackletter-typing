package app

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// IsBossLevel is 10, 20, … 100.
func IsBossLevel(level int) bool {
	return level >= 10 && level <= 100 && level%10 == 0
}

// BossSpec is one literary monster fight.
type BossSpec struct {
	Level       int
	Name        string
	Title       string
	Source      string
	Art         string
	Banner      string
	MinWPM      float64
	MinAccuracy float64
}

// BossFight is live HP/combo/timer state for a fight.
type BossFight struct {
	Spec  BossSpec
	HP    int
	MaxHP int
	Combo int
	Limit time.Duration
}

// BossRoster is the decade bosses, in order.
func BossRoster() []BossSpec {
	levels := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	out := make([]BossSpec, 0, len(levels))
	for _, lv := range levels {
		out = append(out, bossAt(lv))
	}
	return out
}

// BossFor returns the monster for a boss level.
func BossFor(level int) BossSpec {
	if !IsBossLevel(level) {
		return BossSpec{}
	}
	return bossAt(level)
}

func bossAt(level int) BossSpec {
	minWPM := 10 + float64(level)/2 // 15 at 10, 20 at 20, 60 at 100
	minAcc := 0.80 + float64(level)*0.0015
	if minAcc > 0.96 {
		minAcc = 0.96
	}
	base := BossSpec{Level: level, MinWPM: minWPM, MinAccuracy: minAcc}
	switch level {
	case 10:
		base.Name, base.Title, base.Source, base.Art = "Wolfman", "The Wolfman", "folklore & The Wolf Man", artWolfman
	case 20:
		base.Name, base.Title, base.Source, base.Art = "Hyde", "Mr. Hyde", "Robert Louis Stevenson, Strange Case of Dr Jekyll and Mr Hyde", artHyde
	case 30:
		base.Name, base.Title, base.Source, base.Art = "the Monster", "Frankenstein's Monster", "Mary Shelley, Frankenstein", artFrankenstein
	case 40:
		base.Name, base.Title, base.Source, base.Art = "Dracula", "Count Dracula", "Bram Stoker, Dracula", artDracula
	case 50:
		base.Name, base.Title, base.Source, base.Art = "the Whale", "Ahab's Whale", "Herman Melville, Moby-Dick", artWhale
	case 60:
		base.Name, base.Title, base.Source, base.Art = "Grendel", "Grendel", "Beowulf", artGrendel
	case 70:
		base.Name, base.Title, base.Source, base.Art = "the Raven", "The Raven", "Edgar Allan Poe", artRaven
	case 80:
		base.Name, base.Title, base.Source, base.Art = "the Horseman", "the Headless Horseman", "Washington Irving, The Legend of Sleepy Hollow", artHorseman
	case 90:
		base.Name, base.Title, base.Source, base.Art = "the Kraken", "the Kraken", "Tennyson & sailor lore", artKraken
	default:
		base.Name, base.Title, base.Source, base.Art = "Leviathan", "Leviathan", "Job 41 & sailor lore", artLeviathan
		base.Level = 100
	}
	base.Banner = BigBanner(bannerWord(base.Title, base.Name))
	return base
}

func bannerWord(title, name string) string {
	switch {
	case strings.Contains(strings.ToLower(title), "wolf"):
		return "WOLFMAN"
	case strings.Contains(strings.ToLower(title), "hyde"):
		return "HYDE"
	case strings.Contains(strings.ToLower(title), "franken"):
		return "MONSTER"
	case strings.Contains(strings.ToLower(title), "dracula"):
		return "DRACULA"
	case strings.Contains(strings.ToLower(title), "whale"):
		return "WHALE"
	case strings.Contains(strings.ToLower(name), "grendel"):
		return "GRENDEL"
	case strings.Contains(strings.ToLower(title), "raven"):
		return "RAVEN"
	case strings.Contains(strings.ToLower(title), "horseman"):
		return "HORSEMAN"
	case strings.Contains(strings.ToLower(title), "kraken"):
		return "KRAKEN"
	case strings.Contains(strings.ToLower(title), "leviathan"):
		return "LEVIATHAN"
	default:
		return strings.ToUpper(name)
	}
}

// NewBossFight sets HP from the review text length.
func NewBossFight(spec BossSpec, text string) *BossFight {
	n := len([]rune(text))
	if n < 1 {
		n = 1
	}
	hp := n * 2
	return &BossFight{
		Spec:  spec,
		HP:    hp,
		MaxHP: hp,
		Limit: BossTimeLimit(n, spec.MinWPM),
	}
}

// BossTimeLimit is 1.4× the time needed to type the text at the minimum WPM.
func BossTimeLimit(runes int, minWPM float64) time.Duration {
	if minWPM < 1 {
		minWPM = 1
	}
	sec := (float64(runes) / 5.0) / minWPM * 60 * 1.4
	if sec < 25 {
		sec = 25
	}
	return time.Duration(sec * float64(time.Second))
}

// HitDamage is HP lost for one keystroke. Errors deal none.
func HitDamage(correct bool, combo int, wpm, minWPM float64) int {
	if !correct {
		return 0
	}
	dmg := 1
	if combo >= 4 {
		dmg++
	}
	if minWPM > 0 && wpm >= minWPM {
		dmg++
	}
	return dmg
}

// OnKey applies a keystroke to the boss.
func (f *BossFight) OnKey(correct bool, wpm float64) {
	if f == nil {
		return
	}
	f.HP -= HitDamage(correct, f.Combo, wpm, f.Spec.MinWPM)
	if f.HP < 0 {
		f.HP = 0
	}
	if correct {
		f.Combo++
	} else {
		f.Combo = 0
	}
}

// TimedOut reports whether the fight clock has expired.
func (f *BossFight) TimedOut(elapsed time.Duration) bool {
	return f != nil && f.Limit > 0 && elapsed > f.Limit
}

// Remaining is time left on the clock.
func (f *BossFight) Remaining(elapsed time.Duration) time.Duration {
	if f == nil {
		return 0
	}
	left := f.Limit - elapsed
	if left < 0 {
		return 0
	}
	return left
}

// Evaluate decides win/lose after the fight stops.
func (f *BossFight) Evaluate(st SessionStats, finished bool, elapsed time.Duration) (win bool, reason string) {
	if f == nil {
		return false, "fled"
	}
	if f.TimedOut(elapsed) {
		return false, "time"
	}
	if !finished && f.HP > 0 {
		return false, "fled"
	}
	if st.GrossWPM < f.Spec.MinWPM {
		return false, "slow"
	}
	if st.Accuracy < f.Spec.MinAccuracy {
		return false, "errors"
	}
	if f.HP > 0 {
		return false, "alive"
	}
	return true, ""
}

// FailMessage is kid-safe copy for a loss.
func (f *BossFight) FailMessage(reason string) string {
	name := "the boss"
	if f != nil && f.Spec.Name != "" {
		name = f.Spec.Name
	}
	switch reason {
	case "time":
		return fmt.Sprintf("Time is up — %s still stands. Try again!", name)
	case "slow":
		return fmt.Sprintf("Not fast enough — %s got away. Try again!", name)
	case "errors":
		return fmt.Sprintf("Too many misses — %s still stands. Try again!", name)
	case "alive":
		return fmt.Sprintf("%s took the hits and lived. Try again!", strings.ToUpper(name[:1])+name[1:])
	default:
		return fmt.Sprintf("%s still stands. Try again!", strings.ToUpper(name[:1])+name[1:])
	}
}

// WinMessage is kid-safe copy for a win.
func (f *BossFight) WinMessage() string {
	if f == nil {
		return "You did it!"
	}
	return fmt.Sprintf("%s falls!", f.Spec.Title)
}

// BuildBossText is a longer review of recent drills, nonsense, and passages.
func BuildBossText(level int, passages []Passage, rng *rand.Rand) string {
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	target := 100 + (level/10)*40
	var parts []string
	drills := LetterDrills()
	nDrills := 2 + level/25
	for i := 0; i < nDrills && len(drills) > 0; i++ {
		parts = append(parts, Generate(drills[i%len(drills)], rng, 28))
	}
	parts = append(parts, GenerateNonsense(KeysForLevel(level), rng, 40))
	maxTier := 1
	if level >= 30 {
		maxTier = 2
	}
	if level >= 50 {
		maxTier = 3
	}
	for _, p := range SentencePassages(passages) {
		if p.Tier <= maxTier && p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	for _, p := range LiteraturePassages(passages) {
		if p.Tier <= maxTier && p.Text != "" {
			parts = append(parts, p.Text)
		}
	}
	text := strings.Join(parts, " ")
	for len([]rune(text)) < target {
		text += " " + GenerateNonsense(KeysForLevel(level), rng, 24)
	}
	return strings.Join(strings.Fields(text), " ")
}

// FormatClock renders a countdown as M:SS.
func FormatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds() + 0.5)
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}
