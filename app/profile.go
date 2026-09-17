package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	xpPerLevelUnit = 80
	maxLevel       = 100
)

// Score is a personal best for one drill or passage.
type Score struct {
	WPM      float64 `json:"wpm"`
	Accuracy float64 `json:"accuracy"`
	XP       int     `json:"xp"`
}

// Profile is one kid's saved progress.
type Profile struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Level               int              `json:"level"`
	XPIntoLevel         int              `json:"xpIntoLevel"`
	TotalXP             int              `json:"totalXP"`
	TotalTestsCompleted int              `json:"totalTestsCompleted"`
	HighScores          map[string]Score `json:"highScores"`
	KeyStats            map[string]Stat  `json:"keyStats"`
	BigramStats         map[string]Stat  `json:"bigramStats"`
	Journey             map[string]int   `json:"journey"`
	BossesBeaten        []int            `json:"bossesBeaten"`
	CreatedAt           time.Time        `json:"createdAt"`
	LastPlayedAt        time.Time        `json:"lastPlayedAt"`
}

// File is the on-disk document.
type File struct {
	LastProfileID string    `json:"lastProfileID"`
	Profiles      []Profile `json:"profiles"`
}

// Store loads and saves profiles under BLACKLETTER_HOME or ~/.blackletter.
type Store struct {
	path     string
	File     File
	okToSave bool
	LoadErr  error
}

func dataDir() string {
	if d := os.Getenv("BLACKLETTER_HOME"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".blackletter"
	}
	return filepath.Join(home, ".blackletter")
}

// OpenStore loads profiles.json, or starts empty if it does not exist.
// A corrupt file is kept and Save is refused so we never clobber it.
func OpenStore() (*Store, error) {
	dir := dataDir()
	path := filepath.Join(dir, "profiles.json")
	s := &Store{path: path, okToSave: true}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s.File); err != nil {
		s.okToSave = false
		s.LoadErr = fmt.Errorf("could not read profiles (file left untouched): %w", err)
		return s, nil
	}
	for i := range s.File.Profiles {
		s.File.Profiles[i].SyncLevel()
		if s.File.Profiles[i].HighScores == nil {
			s.File.Profiles[i].HighScores = map[string]Score{}
		}
		if s.File.Profiles[i].KeyStats == nil {
			s.File.Profiles[i].KeyStats = map[string]Stat{}
		}
		if s.File.Profiles[i].BigramStats == nil {
			s.File.Profiles[i].BigramStats = map[string]Stat{}
		}
		if s.File.Profiles[i].Journey == nil {
			s.File.Profiles[i].Journey = map[string]int{}
		}
	}
	return s, nil
}

// Save writes profiles atomically. No-op if the original file was corrupt.
func (s *Store) Save() error {
	if s == nil || !s.okToSave {
		if s != nil && s.LoadErr != nil {
			return s.LoadErr
		}
		return errors.New("save refused")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.File, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Profiles returns the slice in store order.
func (s *Store) Profiles() []Profile {
	if s == nil {
		return nil
	}
	return s.File.Profiles
}

// Profile returns a copy of the profile with id, or nil.
func (s *Store) Profile(id string) *Profile {
	p := s.profilePtr(id)
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func (s *Store) profilePtr(id string) *Profile {
	if s == nil {
		return nil
	}
	for i := range s.File.Profiles {
		if s.File.Profiles[i].ID == id {
			return &s.File.Profiles[i]
		}
	}
	return nil
}

// Create adds a named profile and selects it.
func (s *Store) Create(name string) (*Profile, error) {
	name = strings.TrimSpace(name)
	if err := ValidateName(name); err != nil {
		return nil, err
	}
	now := time.Now()
	p := Profile{
		ID:           newID(),
		Name:         name,
		Level:        1,
		HighScores:   map[string]Score{},
		KeyStats:     map[string]Stat{},
		BigramStats:  map[string]Stat{},
		Journey:      map[string]int{},
		BossesBeaten: []int{},
		CreatedAt:    now,
		LastPlayedAt: now,
	}
	s.File.Profiles = append(s.File.Profiles, p)
	s.File.LastProfileID = p.ID
	return s.Profile(p.ID), nil
}

// Select records the last-used profile.
func (s *Store) Select(id string) error {
	if s.profilePtr(id) == nil {
		return errors.New("no such profile")
	}
	s.File.LastProfileID = id
	return nil
}

// LastProfile is the previously selected profile, if it still exists.
func (s *Store) LastProfile() *Profile {
	if s == nil || s.File.LastProfileID == "" {
		return nil
	}
	return s.Profile(s.File.LastProfileID)
}

// Award applies XP, high scores, and miss stats after a session.
func (s *Store) Award(id, scoreKey string, xp int, score Score, keys map[rune]Stat, bigrams map[string]Stat) *Profile {
	p := s.profilePtr(id)
	if p == nil {
		return nil
	}
	p.TotalXP += xp
	p.TotalTestsCompleted++
	p.LastPlayedAt = time.Now()
	p.SyncLevel()
	if p.HighScores == nil {
		p.HighScores = map[string]Score{}
	}
	if prev, ok := p.HighScores[scoreKey]; !ok || BetterScore(score, prev) {
		p.HighScores[scoreKey] = score
	}
	if p.KeyStats == nil {
		p.KeyStats = map[string]Stat{}
	}
	for r, st := range keys {
		k := string(r)
		p.KeyStats[k] = MergeStat(p.KeyStats[k], st)
	}
	if p.BigramStats == nil {
		p.BigramStats = map[string]Stat{}
	}
	for bg, st := range bigrams {
		p.BigramStats[bg] = MergeStat(p.BigramStats[bg], st)
	}
	cp := *p
	return &cp
}

// MarkJourney records a successful Play step.
func (s *Store) MarkJourney(id, stepID string) *Profile {
	p := s.profilePtr(id)
	if p == nil || stepID == "" {
		return s.Profile(id)
	}
	if p.Journey == nil {
		p.Journey = map[string]int{}
	}
	p.Journey[stepID]++
	cp := *p
	return &cp
}

// HasBeatBoss reports whether this decade boss is already down.
func (p Profile) HasBeatBoss(level int) bool {
	for _, lv := range p.BossesBeaten {
		if lv == level {
			return true
		}
	}
	return false
}

// MarkBoss records a boss win.
func (s *Store) MarkBoss(id string, level int) *Profile {
	p := s.profilePtr(id)
	if p == nil {
		return nil
	}
	if !p.HasBeatBoss(level) {
		p.BossesBeaten = append(p.BossesBeaten, level)
	}
	cp := *p
	return &cp
}

// HallOfFame returns profiles sorted by level, then XP, then name.
func (s *Store) HallOfFame() []Profile {
	out := append([]Profile(nil), s.File.Profiles...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Level != out[j].Level {
			return out[i].Level > out[j].Level
		}
		if out[i].TotalXP != out[j].TotalXP {
			return out[i].TotalXP > out[j].TotalXP
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

// SyncLevel recomputes Level and XPIntoLevel from TotalXP.
func (p *Profile) SyncLevel() {
	p.Level, p.XPIntoLevel, _ = LevelForTotalXP(p.TotalXP)
}

// XPToReachLevel is total XP at the start of a 1-based level.
func XPToReachLevel(level int) int {
	if level <= 1 {
		return 0
	}
	if level > maxLevel {
		level = maxLevel
	}
	total := 0
	for l := 1; l < level; l++ {
		total += xpPerLevelUnit * l
	}
	return total
}

// SkipToLevel sets the profile at the beginning of the given level.
func (s *Store) SkipToLevel(id string, level int) *Profile {
	p := s.profilePtr(id)
	if p == nil {
		return nil
	}
	if level < 1 {
		level = 1
	}
	if level > maxLevel {
		level = maxLevel
	}
	p.TotalXP = XPToReachLevel(level)
	p.SyncLevel()
	cp := *p
	return &cp
}

// XPNeeded is XP required to finish the current level.
func (p Profile) XPNeeded() int {
	_, _, need := LevelForTotalXP(p.TotalXP)
	return need
}

// KeyStatMap converts string-keyed JSON stats to runes.
func (p Profile) KeyStatMap() map[rune]Stat {
	out := make(map[rune]Stat, len(p.KeyStats))
	for k, st := range p.KeyStats {
		rs := []rune(k)
		if len(rs) != 1 {
			continue
		}
		out[rs[0]] = st
	}
	return out
}

// LevelForTotalXP returns current level (1-based), XP into that level, and XP needed to finish it.
func LevelForTotalXP(total int) (level, into, needed int) {
	if total < 0 {
		total = 0
	}
	level = 1
	remaining := total
	for level < maxLevel {
		needed = xpPerLevelUnit * level
		if remaining < needed {
			return level, remaining, needed
		}
		remaining -= needed
		level++
	}
	needed = xpPerLevelUnit * maxLevel
	return maxLevel, remaining, needed
}

// MaxUnlockedTier is the highest passage tier a level may play.
// Level 1: drills only (0). Level 2: tier 1. Level 3–4: tier 2. Level 5+: tier 3.
func MaxUnlockedTier(level int) int {
	switch {
	case level >= 5:
		return 3
	case level >= 3:
		return 2
	case level >= 2:
		return 1
	default:
		return 0
	}
}

// ComputeXP converts a session into experience points.
// literatureBonus is a 2× multiplier when a literature passage is finished at the pass bar.
func ComputeXP(st SessionStats, finished, literatureBonus bool) int {
	if st.Attempts == 0 {
		return 0
	}
	wpmFactor := 1 + math.Min(st.GrossWPM, 80)/80
	raw := float64(st.Correct) * st.Accuracy * wpmFactor
	if finished && st.Accuracy >= 1 {
		raw *= 1.15
	}
	if !finished {
		raw *= 0.5
	}
	if literatureBonus && finished && st.Accuracy >= JourneyPassAccuracy {
		raw *= 2
	}
	xp := int(math.Round(raw))
	if xp < 1 {
		xp = 1
	}
	return xp
}

// BetterScore prefers higher XP, then accuracy, then WPM.
func BetterScore(a, b Score) bool {
	if a.XP != b.XP {
		return a.XP > b.XP
	}
	if a.Accuracy != b.Accuracy {
		return a.Accuracy > b.Accuracy
	}
	return a.WPM > b.WPM
}

// ValidateName allows 2–16 letters and spaces.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	n := len([]rune(name))
	if n < 2 {
		return errors.New("name needs at least 2 letters")
	}
	if n > 16 {
		return errors.New("name is too long (16 letters max)")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' {
			return errors.New("use letters and spaces only")
		}
	}
	return nil
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
