package app

import (
	"math/rand/v2"
	"strings"
	"unicode"
)

const (
	drillTargetLen = 48
	homeRow        = "asdfjkl;"
)

// Drill is a named curriculum unit. Text is generated per session.
type Drill struct {
	ID       string
	Title    string
	Desc     string
	MinLevel int
	Keys     []rune
	Words    bool
}

// Curriculum is the progressive home-row lesson list.
var Curriculum = []Drill{
	{ID: "home-fj", Title: "Index bumps", Desc: "f and j — feel the little bumps", MinLevel: 1, Keys: []rune("fj")},
	{ID: "home-dk", Title: "Middle fingers", Desc: "d and k sit next to the bumps", MinLevel: 1, Keys: []rune("dk")},
	{ID: "home-sl", Title: "Ring fingers", Desc: "s and l — stretch a little", MinLevel: 1, Keys: []rune("sl")},
	{ID: "home-a;", Title: "Pinkies", Desc: "a and ; — the home-row ends", MinLevel: 1, Keys: []rune("a;")},
	{ID: "home-row", Title: "Full home row", Desc: "asdf jkl; — rest your fingers here", MinLevel: 1, Keys: []rune("asdfjkl;")},
	{ID: "home-gh", Title: "Reach in", Desc: "g and h — index fingers stretch in", MinLevel: 1, Keys: []rune("gh")},
	{ID: "top-eio", Title: "Top-row vowels", Desc: "e i o — reach up from home row", MinLevel: 1, Keys: []rune("eio")},
}

var easyWords = []string{
	"a", "as", "add", "sad", "lad", "all", "fall", "ask", "flask",
	"dad", "fad", "jak", "lass", "salad", "flask", "falls", "all",
	"ja", "ka", "la", "fa", "ska", "alfalfa", "salsa",
}

// UnlockedDrills are curriculum items the profile's level may play.
func UnlockedDrills(level int) []Drill {
	var out []Drill
	for _, d := range Curriculum {
		if level >= d.MinLevel {
			out = append(out, d)
		}
	}
	return out
}

// Generate builds a fresh drill string from the allowed keys.
func Generate(d Drill, rng *rand.Rand, n int) string {
	if n <= 0 {
		n = drillTargetLen
	}
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	if d.Words {
		return generateWords(d.Keys, rng, n)
	}
	return generateKeys(d.Keys, rng, n)
}

func generateKeys(keys []rune, rng *rand.Rand, n int) string {
	if len(keys) == 0 {
		keys = []rune(homeRow)
	}
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		group := 2 + rng.IntN(3)
		for i := 0; i < group; i++ {
			b.WriteRune(keys[rng.IntN(len(keys))])
		}
	}
	return strings.TrimSpace(b.String())
}

func generateWords(keys []rune, rng *rand.Rand, n int) string {
	allowed := make(map[rune]bool, len(keys))
	for _, r := range keys {
		allowed[r] = true
	}
	var pool []string
	for _, w := range easyWords {
		if wordAllowed(w, allowed) {
			pool = append(pool, w)
		}
	}
	if len(pool) == 0 {
		return generateKeys(keys, rng, n)
	}
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(pool[rng.IntN(len(pool))])
	}
	return strings.TrimSpace(b.String())
}

func wordAllowed(w string, allowed map[rune]bool) bool {
	for _, r := range w {
		if !allowed[r] {
			return false
		}
	}
	return true
}

// GenerateAdaptive builds a remediation drill: ~60% weak keys, 40% home row.
func GenerateAdaptive(weak []rune, rng *rand.Rand, n int) string {
	if n <= 0 {
		n = drillTargetLen
	}
	if rng == nil {
		rng = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	home := []rune(homeRow)
	if len(weak) == 0 {
		return generateKeys(home, rng, n)
	}
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		group := 2 + rng.IntN(3)
		for i := 0; i < group; i++ {
			var r rune
			if rng.IntN(10) < 6 {
				r = weak[rng.IntN(len(weak))]
			} else {
				r = home[rng.IntN(len(home))]
			}
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// AdaptiveFromProfile picks weak keys from saved stats and generates a drill.
func AdaptiveFromProfile(p Profile, rng *rand.Rand) (title, desc, text, id string) {
	weak := WeakKeys(p.KeyStatMap(), 5, 0.20)
	if len(weak) == 0 {
		d := Curriculum[0]
		return "Recommended", "A warm-up on your home-row bumps", Generate(d, rng, drillTargetLen), "drill:recommended"
	}
	shown := make([]string, 0, len(weak))
	for _, r := range weak {
		if unicode.IsPrint(r) {
			shown = append(shown, string(r))
		}
	}
	desc = "Practice your tricky keys: " + strings.Join(shown, " ")
	return "Recommended", desc, GenerateAdaptive(weak, rng, drillTargetLen), "drill:recommended"
}

// OnlyAllowed reports whether text uses only the given keys plus spaces.
func OnlyAllowed(text string, keys []rune) bool {
	ok := make(map[rune]bool, len(keys)+1)
	ok[' '] = true
	for _, r := range keys {
		ok[r] = true
	}
	for _, r := range text {
		if !ok[r] {
			return false
		}
	}
	return true
}
