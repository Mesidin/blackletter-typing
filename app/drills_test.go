package app

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode"
)

func TestGenerateUsesOnlyAllowedKeys(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, d := range Curriculum {
		text := Generate(d, rng, 60)
		if text == "" {
			t.Fatalf("%s empty", d.ID)
		}
		if !OnlyAllowed(text, append(d.Keys, ' ')) {
			t.Fatalf("%s used extra keys: %q", d.ID, text)
		}
	}
}

func TestGenerateAdaptiveBiasedTowardWeak(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	weak := []rune{'p', 'q'}
	var p, q, other int
	for i := 0; i < 50; i++ {
		text := GenerateAdaptive(weak, rng, 80)
		for _, r := range text {
			if unicode.IsSpace(r) {
				continue
			}
			switch r {
			case 'p':
				p++
			case 'q':
				q++
			default:
				other++
			}
		}
	}
	weakN := p + q
	if weakN < other {
		t.Fatalf("expected weak majority, weak=%d other=%d", weakN, other)
	}
	if p == 0 || q == 0 {
		t.Fatalf("both weak keys should appear p=%d q=%d", p, q)
	}
}

func TestUnlockedDrills(t *testing.T) {
	l1 := UnlockedDrills(1)
	for _, d := range l1 {
		if d.MinLevel > 1 {
			t.Fatalf("level 1 got %s", d.ID)
		}
	}
	letters := LetterDrills()
	if len(letters) < 7 {
		t.Fatalf("expected a full home-row path, got %d", len(letters))
	}
}

func TestAdaptiveFromProfileFallsBack(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	title, _, text, id := AdaptiveFromProfile(Profile{}, rng)
	if title == "" || text == "" || !strings.HasPrefix(id, "drill:") {
		t.Fatalf("fallback %q %q %q", title, text, id)
	}
}
