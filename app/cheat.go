package app

import (
	"fmt"
	"strconv"
	"strings"
)

// Konami-style code, armed on menus only (never during a lesson).
// ↑ ↑ ↓ ↓ ← → ← → B A  → prompt for a level to skip to
var cheatSeq = []string{"up", "up", "down", "down", "left", "right", "left", "right", "b", "a"}

const cheatBufMax = 10

func feedCheat(buf []string, key string) (next []string, hit bool) {
	if key == "" {
		return buf, false
	}
	next = append(buf, key)
	if len(next) > cheatBufMax {
		next = append([]string(nil), next[len(next)-cheatBufMax:]...)
	}
	if matchSeq(next, cheatSeq) {
		return nil, true
	}
	return next, false
}

func matchSeq(buf, seq []string) bool {
	if len(buf) < len(seq) {
		return false
	}
	tail := buf[len(buf)-len(seq):]
	for i := range seq {
		if tail[i] != seq[i] {
			return false
		}
	}
	return true
}

func parseCheatLevel(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("type a number")
	}
	if n < 1 || n > maxLevel {
		return 0, fmt.Errorf("level must be 1–%d", maxLevel)
	}
	return n, nil
}
