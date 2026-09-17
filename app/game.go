package app

import (
	"time"
	"unicode"
)

// Stat counts attempts and misses for a key or bigram.
type Stat struct {
	Attempts int `json:"attempts"`
	Misses   int `json:"misses"`
}

// MissRate is misses/attempts, or 0 if unused.
func (s Stat) MissRate() float64 {
	if s.Attempts == 0 {
		return 0
	}
	return float64(s.Misses) / float64(s.Attempts)
}

func (s Stat) hit(ok bool) Stat {
	s.Attempts++
	if !ok {
		s.Misses++
	}
	return s
}

// MergeStat adds b into a.
func MergeStat(a, b Stat) Stat {
	return Stat{Attempts: a.Attempts + b.Attempts, Misses: a.Misses + b.Misses}
}

// Session is a single typing run against a target string.
type Session struct {
	Target []rune
	Typed  []rune
	Missed []bool
	Pos    int

	Now func() time.Time

	started  time.Time
	finished time.Time
	attempts int
	correct  int
	keyStats map[rune]Stat
	bigrams  map[string]Stat
}

// SessionStats is a snapshot of speed and accuracy.
type SessionStats struct {
	Attempts int
	Correct  int
	Accuracy float64
	GrossWPM float64
	NetWPM   float64
	Elapsed  time.Duration
	Done     bool
}

// NewSession starts an untimed session; the clock begins on the first key.
func NewSession(target string) *Session {
	t := []rune(target)
	return &Session{
		Target:   t,
		Missed:   make([]bool, len(t)),
		Now:      time.Now,
		keyStats: make(map[rune]Stat),
		bigrams:  make(map[string]Stat),
	}
}

// Handle records a printable keystroke against the current target rune.
func (s *Session) Handle(r rune) {
	if s == nil || s.Done() || s.Pos >= len(s.Target) {
		return
	}
	if s.started.IsZero() {
		s.started = s.now()
	}
	expected := s.Target[s.Pos]
	ok := r == expected
	s.attempts++
	if ok {
		s.correct++
	} else {
		s.Missed[s.Pos] = true
	}
	s.keyStats[expected] = s.keyStats[expected].hit(ok)
	if s.Pos > 0 {
		bg := string([]rune{s.Target[s.Pos-1], expected})
		s.bigrams[bg] = s.bigrams[bg].hit(ok)
	}
	if s.Pos == len(s.Typed) {
		s.Typed = append(s.Typed, r)
	} else {
		s.Typed[s.Pos] = r
	}
	s.Pos++
	if s.Pos == len(s.Target) {
		s.finished = s.now()
	}
}

// Backspace moves the caret back one rune. Miss history is kept.
func (s *Session) Backspace() {
	if s == nil || s.Pos == 0 {
		return
	}
	s.Pos--
	if len(s.Typed) > s.Pos {
		s.Typed = s.Typed[:s.Pos]
	}
	if !s.finished.IsZero() {
		s.finished = time.Time{}
	}
}

// Done reports whether every target rune has been typed.
func (s *Session) Done() bool {
	return s != nil && len(s.Target) > 0 && s.Pos >= len(s.Target)
}

// Started reports whether the clock has begun.
func (s *Session) Started() bool {
	return s != nil && !s.started.IsZero()
}

// Attempts is the count of non-backspace keystrokes.
func (s *Session) Attempts() int { return s.attempts }

// Correct is the count of matching keystrokes.
func (s *Session) Correct() int { return s.correct }

// KeyStats returns a copy of per-expected-key counters.
func (s *Session) KeyStats() map[rune]Stat {
	out := make(map[rune]Stat, len(s.keyStats))
	for k, v := range s.keyStats {
		out[k] = v
	}
	return out
}

// BigramStats returns a copy of per-bigram counters.
func (s *Session) BigramStats() map[string]Stat {
	out := make(map[string]Stat, len(s.bigrams))
	for k, v := range s.bigrams {
		out[k] = v
	}
	return out
}

// NextRune is the rune under the caret, or 0 if done/empty.
func (s *Session) NextRune() rune {
	if s == nil || s.Pos >= len(s.Target) {
		return 0
	}
	return s.Target[s.Pos]
}

// Elapsed is time since the first keystroke.
func (s *Session) Elapsed() time.Duration {
	if s == nil || s.started.IsZero() {
		return 0
	}
	end := s.now()
	if !s.finished.IsZero() {
		end = s.finished
	}
	d := end.Sub(s.started)
	if d < 0 {
		return 0
	}
	return d
}

// Stats snapshots WPM and accuracy using the current clock.
func (s *Session) Stats() SessionStats {
	if s == nil {
		return SessionStats{}
	}
	acc := 0.0
	if s.attempts > 0 {
		acc = float64(s.correct) / float64(s.attempts)
	}
	elapsed := s.Elapsed()
	wpm := 0.0
	if minutes := elapsed.Minutes(); minutes > 0 {
		wpm = (float64(s.attempts) / 5.0) / minutes
	}
	return SessionStats{
		Attempts: s.attempts,
		Correct:  s.correct,
		Accuracy: acc,
		GrossWPM: wpm,
		NetWPM:   wpm * acc,
		Elapsed:  elapsed,
		Done:     s.Done(),
	}
}

func (s *Session) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// WeakKeys returns expected keys with enough samples and a high miss rate.
func WeakKeys(stats map[rune]Stat, minAttempts int, missRate float64) []rune {
	var out []rune
	for r, st := range stats {
		if unicode.IsSpace(r) {
			continue
		}
		if st.Attempts >= minAttempts && st.MissRate() >= missRate {
			out = append(out, r)
		}
	}
	return out
}

// WeakBigrams returns bigrams with enough samples and a high miss rate.
func WeakBigrams(stats map[string]Stat, minAttempts int, missRate float64) []string {
	var out []string
	for bg, st := range stats {
		if st.Attempts >= minAttempts && st.MissRate() >= missRate {
			out = append(out, bg)
		}
	}
	return out
}
