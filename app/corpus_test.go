package app

import (
	"testing"
	"testing/fstest"
)

func TestLoadCorpus(t *testing.T) {
	fsys := fstest.MapFS{
		"literature/alice.json": {Data: []byte(`{"id":"alice-01","title":"Alice","source":"Carroll","category":"literature","tier":2,"text":"Alice was beginning to get very tired."}`)},
		"poetry/hope.json":      {Data: []byte(`{"id":"hope","title":"Hope","source":"Dickinson","category":"poetry","tier":3,"text":"Hope is the thing with feathers."}`)},
		"skip.txt":              {Data: []byte("not json")},
	}
	ps, err := LoadCorpus(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("got %d", len(ps))
	}
	if ps[0].ID != "alice-01" {
		t.Fatalf("sort by tier, first=%s", ps[0].ID)
	}
}

func TestLoadCorpusRejectsEmpty(t *testing.T) {
	fsys := fstest.MapFS{
		"bad.json": {Data: []byte(`{"id":"x","text":""}`)},
	}
	if _, err := LoadCorpus(fsys); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnlockedPassages(t *testing.T) {
	all := []Passage{
		{ID: "a", Tier: 1},
		{ID: "b", Tier: 2},
		{ID: "c", Tier: 3},
	}
	if n := len(UnlockedPassages(all, 1)); n != 0 {
		t.Fatalf("level 1: %d", n)
	}
	if n := len(UnlockedPassages(all, 2)); n != 1 {
		t.Fatalf("level 2: %d", n)
	}
	if n := len(UnlockedPassages(all, 5)); n != 3 {
		t.Fatalf("level 5: %d", n)
	}
}

func TestRecommendUnplayedThenWeakest(t *testing.T) {
	all := []Passage{
		{ID: "t1", Tier: 1, Title: "one"},
		{ID: "t2", Tier: 2, Title: "two"},
		{ID: "t2b", Tier: 2, Title: "two-b"},
	}
	p := Profile{Level: 3, HighScores: map[string]Score{}}
	got, ok := Recommend(all, p)
	if !ok || got.ID != "t2" && got.ID != "t2b" {
		t.Fatalf("want a tier-2 unplayed, got %+v ok=%v", got, ok)
	}
	if got.Tier != 2 {
		t.Fatalf("tier %d", got.Tier)
	}

	p.HighScores["passage:t2"] = Score{Accuracy: 0.99}
	p.HighScores["passage:t2b"] = Score{Accuracy: 0.99}
	got, ok = Recommend(all, p)
	if !ok || got.ID != "t1" {
		t.Fatalf("unplayed t1 next, got %s", got.ID)
	}

	p.HighScores["passage:t1"] = Score{Accuracy: 0.5}
	p.HighScores["passage:t2"] = Score{Accuracy: 0.9}
	p.HighScores["passage:t2b"] = Score{Accuracy: 0.8}
	got, ok = Recommend(all, p)
	if !ok || got.ID != "t1" {
		t.Fatalf("lowest accuracy should be t1, got %s", got.ID)
	}
}

func TestRecommendEmpty(t *testing.T) {
	if _, ok := Recommend(nil, Profile{Level: 5}); ok {
		t.Fatal("empty corpus")
	}
}
