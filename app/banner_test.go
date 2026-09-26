package app

import (
	"strings"
	"testing"
)

func TestBigBanner(t *testing.T) {
	b := BigBanner("WOLFMAN")
	lines := strings.Split(b, "\n")
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines in BigBanner, got %d", len(lines))
	}
	if !strings.Contains(b, "█") {
		t.Errorf("expected block characters in BigBanner, got %q", b)
	}
}
