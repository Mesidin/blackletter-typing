package main

import (
	"testing"

	"blackletter/app"
)

func TestEmbeddedPassagesLoad(t *testing.T) {
	ps, err := app.LoadCorpus(textsFS)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) < 8 {
		t.Fatalf("expected a full starter corpus, got %d", len(ps))
	}
	seen := map[string]bool{}
	for _, p := range ps {
		if p.Text == "" || p.ID == "" {
			t.Fatalf("empty passage %+v", p)
		}
		if seen[p.ID] {
			t.Fatalf("duplicate id %s", p.ID)
		}
		seen[p.ID] = true
	}
	for _, id := range []string{"alice-01", "frank-01", "dickinson-hope", "poe-raven", "psalm-23", "cat-mat"} {
		if !seen[id] {
			t.Fatalf("missing %s", id)
		}
	}
}
