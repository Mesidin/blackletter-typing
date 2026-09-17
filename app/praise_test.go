package app

import "testing"

func TestPraiseNeverHarsh(t *testing.T) {
	msgs := []string{
		Praise(1, true),
		Praise(0.9, true),
		Praise(0.5, true),
		Praise(0.2, false),
	}
	for _, m := range msgs {
		if m == "" {
			t.Fatal("empty praise")
		}
	}
	if Praise(1, true) != "PERFECT! You nailed every letter!" {
		t.Fatalf("perfect copy: %q", Praise(1, true))
	}
}
