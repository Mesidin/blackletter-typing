package app

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// JourneyPassAccuracy is the bar for a Play step to count as done.
const JourneyPassAccuracy = 0.80

// JourneyKind is a stage in the Play path.
type JourneyKind int

const (
	JourneyLetters JourneyKind = iota
	JourneyNonsense
	JourneySentence
)

// JourneyStep is one required chunk of the Play journey at a level.
type JourneyStep struct {
	ID    string
	Kind  JourneyKind
	Title string
	Desc  string
	Need  int
	Drill Drill
}

// LetterDrills is the home-row curriculum without word lists.
func LetterDrills() []Drill {
	var out []Drill
	for _, d := range Curriculum {
		if !d.Words {
			out = append(out, d)
		}
	}
	return out
}

func nonsenseDrill(level int) Drill {
	return Drill{
		ID:    "nonsense",
		Title: "Nonsense words",
		Desc:  "Fake little words — just mash the right keys",
		Keys:  KeysForLevel(level),
	}
}

// JourneySteps is the enforced Play path for a level:
// lots of letter drills, then nonsense words, then sentences once unlocked.
func JourneySteps(level int) []JourneyStep {
	if level < 1 {
		level = 1
	}
	if IsBossLevel(level) {
		return nil
	}
	tag := fmt.Sprintf("L%d", level)
	needLetters := 2
	if level >= 2 {
		needLetters = 1
	}
	var steps []JourneyStep
	for _, d := range LetterDrills() {
		title := d.Title
		if level >= 2 {
			title = d.Title + " · review"
		}
		steps = append(steps, JourneyStep{
			ID:    tag + ":" + d.ID,
			Kind:  JourneyLetters,
			Title: title,
			Desc:  d.Desc,
			Need:  needLetters,
			Drill: d,
		})
	}
	steps = append(steps, JourneyStep{
		ID:    tag + ":nonsense",
		Kind:  JourneyNonsense,
		Title: "Nonsense words",
		Desc:  "Short made-up words from the keys you know",
		Need:  3,
		Drill: nonsenseDrill(level),
	})
	if MaxUnlockedTier(level) >= 1 {
		steps = append(steps, JourneyStep{
			ID:    tag + ":sentences",
			Kind:  JourneySentence,
			Title: "Short sentences",
			Desc:  "Real words in a row",
			Need:  2,
		})
	}
	return steps
}

// NextJourneyStep is the first incomplete Play step, or a bonus round if the path is done.
func NextJourneyStep(p Profile) (step JourneyStep, done int, bonus bool) {
	steps := JourneySteps(p.Level)
	for _, s := range steps {
		n := p.JourneyCount(s.ID)
		if n < s.Need {
			return s, n, false
		}
	}
	return bonusStep(p.Level), p.JourneyCount(bonusStep(p.Level).ID), true
}

func bonusStep(level int) JourneyStep {
	tag := fmt.Sprintf("L%d:bonus", level)
	if MaxUnlockedTier(level) >= 1 {
		return JourneyStep{
			ID:    tag + "-sentence",
			Kind:  JourneySentence,
			Title: "Bonus sentences",
			Desc:  "Path complete — extra sentence practice",
			Need:  1,
		}
	}
	return JourneyStep{
		ID:    tag + "-nonsense",
		Kind:  JourneyNonsense,
		Title: "Bonus nonsense words",
		Desc:  "Path complete — extra word practice",
		Need:  1,
		Drill: nonsenseDrill(level),
	}
}

// PlayDesc is the main-menu subtitle for Play.
func PlayDesc(p Profile) string {
	if IsBossLevel(p.Level) && !p.HasBeatBoss(p.Level) {
		b := BossFor(p.Level)
		return fmt.Sprintf("BOSS · %s  (%d WPM / %.0f%%)", b.Title, int(b.MinWPM), b.MinAccuracy*100)
	}
	step, done, bonus := NextJourneyStep(p)
	if bonus {
		return "Path complete — bonus: " + step.Title
	}
	return fmt.Sprintf("Next: %s  %d/%d", step.Title, done, step.Need)
}

// JourneyCount is successful Play completions for a step.
func (p Profile) JourneyCount(id string) int {
	if p.Journey == nil {
		return 0
	}
	return p.Journey[id]
}

// KeysForLevel is the letter pool for nonsense drills.
func KeysForLevel(level int) []rune {
	switch {
	case level >= 3:
		return []rune("asdfjkl;gheio")
	case level >= 2:
		return []rune("asdfjkl;gh")
	default:
		return []rune(homeRow)
	}
}

// GenerateNonsense builds space-separated 3–5 letter fake words.
func GenerateNonsense(keys []rune, rng *rand.Rand, n int) string {
	if n <= 0 {
		n = drillTargetLen
	}
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	if len(keys) == 0 {
		keys = []rune(homeRow)
	}
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		wordLen := 3 + rng.IntN(3)
		for i := 0; i < wordLen; i++ {
			b.WriteRune(keys[rng.IntN(len(keys))])
		}
	}
	return strings.TrimSpace(b.String())
}
