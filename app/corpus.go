package app

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Passage is a typed excerpt with a difficulty tier.
type Passage struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Category string `json:"category"`
	Tier     int    `json:"tier"`
	Text     string `json:"text"`
	Kind     string `json:"kind"`
}

// ScoreKey is the high-score map key for a passage.
func (p Passage) ScoreKey() string { return "passage:" + p.ID }

// LoadCorpus walks a filesystem for *.json passage files.
func LoadCorpus(fsys fs.FS) ([]Passage, error) {
	var out []Passage
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		var p Passage
		if err := json.Unmarshal(b, &p); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		p.Text = strings.TrimSpace(p.Text)
		if p.ID == "" || p.Text == "" {
			return fmt.Errorf("%s: missing id or text", path)
		}
		if p.Tier < 1 {
			p.Tier = 1
		}
		if p.Category == "" {
			p.Category = "literature"
		}
		if p.Kind == "" {
			p.Kind = "literature"
		}
		out = append(out, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier < out[j].Tier
		}
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Title < out[j].Title
	})
	return out, nil
}

// UnlockedPassages are those at or below the profile's max tier.
func UnlockedPassages(passages []Passage, level int) []Passage {
	max := MaxUnlockedTier(level)
	var out []Passage
	for _, p := range passages {
		if p.Tier <= max {
			out = append(out, p)
		}
	}
	return out
}

// IsSentence is a short Play-path sentence, not a literature passage.
func (p Passage) IsSentence() bool { return p.Kind == "sentence" }

// LiteraturePassages are classic excerpts for Literature mode.
func LiteraturePassages(passages []Passage) []Passage {
	var out []Passage
	for _, p := range passages {
		if !p.IsSentence() {
			out = append(out, p)
		}
	}
	return out
}

// SentencePassages are the short lines used in the Play journey.
func SentencePassages(passages []Passage) []Passage {
	var out []Passage
	for _, p := range passages {
		if p.IsSentence() {
			out = append(out, p)
		}
	}
	return out
}

// NextSentence picks an unlocked Play sentence, preferring ones without a high score.
func NextSentence(passages []Passage, p Profile) (Passage, bool) {
	return Recommend(SentencePassages(passages), p)
}

// Recommend picks the highest-tier unlocked passage without a high score.
// If all have scores, it returns the one with the lowest accuracy.
func Recommend(passages []Passage, p Profile) (Passage, bool) {
	unlocked := UnlockedPassages(passages, p.Level)
	if len(unlocked) == 0 {
		return Passage{}, false
	}
	sort.SliceStable(unlocked, func(i, j int) bool {
		return unlocked[i].Tier > unlocked[j].Tier
	})
	for _, pas := range unlocked {
		if _, ok := p.HighScores[pas.ScoreKey()]; !ok {
			return pas, true
		}
	}
	best := unlocked[0]
	bestAcc := 2.0
	for _, pas := range unlocked {
		sc := p.HighScores[pas.ScoreKey()]
		if sc.Accuracy < bestAcc {
			bestAcc = sc.Accuracy
			best = pas
		}
	}
	return best, true
}
