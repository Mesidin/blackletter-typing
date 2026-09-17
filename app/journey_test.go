package app

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode"
)

func TestJourneyLevel1IsLotsOfDrillsThenNonsense(t *testing.T) {
	steps := JourneySteps(1)
	if len(steps) < 8 {
		t.Fatalf("level 1 should be a long path, got %d steps", len(steps))
	}
	var letters, nonsense, sentences int
	for _, s := range steps {
		switch s.Kind {
		case JourneyLetters:
			letters++
			if s.Need < 2 {
				t.Fatalf("%s should require multiple completions, need=%d", s.ID, s.Need)
			}
		case JourneyNonsense:
			nonsense++
		case JourneySentence:
			sentences++
		}
	}
	if letters < 7 {
		t.Fatalf("letter drills %d", letters)
	}
	if nonsense != 1 {
		t.Fatalf("nonsense stages %d", nonsense)
	}
	if sentences != 0 {
		t.Fatal("level 1 should not include sentences")
	}
	if steps[len(steps)-1].Kind != JourneyNonsense {
		t.Fatal("nonsense should come after the letter drills")
	}
}

func TestJourneyLevel2AddsSentencesAfterNonsense(t *testing.T) {
	steps := JourneySteps(2)
	if steps[len(steps)-1].Kind != JourneySentence {
		t.Fatalf("last step %v", steps[len(steps)-1].Kind)
	}
	foundNonsense := false
	for _, s := range steps {
		if s.Kind == JourneyNonsense {
			foundNonsense = true
		}
		if s.Kind == JourneySentence && !foundNonsense {
			t.Fatal("sentences should follow nonsense words")
		}
	}
}

func TestNextJourneyStepAdvancesAfterSuccesses(t *testing.T) {
	p := Profile{Level: 1, Journey: map[string]int{}}
	first, done, bonus := NextJourneyStep(p)
	if bonus || first.Kind != JourneyLetters || done != 0 {
		t.Fatalf("first step %+v done=%d bonus=%v", first, done, bonus)
	}
	p.Journey[first.ID] = first.Need
	second, _, _ := NextJourneyStep(p)
	if second.ID == first.ID {
		t.Fatal("should advance after meeting Need")
	}
}

func TestNextJourneyStepBonusWhenPathComplete(t *testing.T) {
	p := Profile{Level: 1, Journey: map[string]int{}}
	for _, s := range JourneySteps(1) {
		p.Journey[s.ID] = s.Need
	}
	step, _, bonus := NextJourneyStep(p)
	if !bonus {
		t.Fatal("expected bonus round")
	}
	if step.Kind != JourneyNonsense {
		t.Fatalf("level 1 bonus should be nonsense, got %v", step.Kind)
	}
}

func TestPlayDesc(t *testing.T) {
	p := Profile{Level: 1, Journey: map[string]int{}}
	d := PlayDesc(p)
	if !strings.Contains(d, "Next:") || !strings.Contains(d, "0/") {
		t.Fatalf("desc %q", d)
	}
	boss := PlayDesc(Profile{Level: 10})
	if !strings.Contains(boss, "BOSS") || !strings.Contains(boss, "Wolfman") {
		t.Fatalf("boss desc %q", boss)
	}
	beaten := PlayDesc(Profile{Level: 10, BossesBeaten: []int{10}})
	if strings.Contains(beaten, "BOSS") {
		t.Fatalf("beaten boss still showing: %q", beaten)
	}
}

func TestJourneyStepsEmptyOnBossLevel(t *testing.T) {
	if steps := JourneySteps(10); len(steps) != 0 {
		t.Fatalf("boss level should not run the normal path, got %d", len(steps))
	}
}

func TestGenerateNonsense(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 10))
	keys := []rune("asdfjkl;")
	text := GenerateNonsense(keys, rng, 40)
	if text == "" || !strings.Contains(text, " ") {
		t.Fatalf("expected words, got %q", text)
	}
	if !OnlyAllowed(text, append(keys, ' ')) {
		t.Fatalf("extra keys in %q", text)
	}
	for _, w := range strings.Fields(text) {
		n := len([]rune(w))
		if n < 3 || n > 5 {
			t.Fatalf("word length %d in %q", n, w)
		}
		for _, r := range w {
			if unicode.IsSpace(r) {
				t.Fatal("space inside word")
			}
		}
	}
}

func TestNextSentencePrefersUnplayed(t *testing.T) {
	all := []Passage{
		{ID: "s1", Kind: "sentence", Tier: 1, Title: "one", Text: "hi"},
		{ID: "s2", Kind: "sentence", Tier: 1, Title: "two", Text: "yo"},
		{ID: "lit", Kind: "literature", Tier: 1, Title: "book", Text: "nope"},
	}
	p := Profile{Level: 2, HighScores: map[string]Score{"passage:s1": {Accuracy: 1}}}
	got, ok := NextSentence(all, p)
	if !ok || got.ID != "s2" {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}
