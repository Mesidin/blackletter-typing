package app

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Palette is the semantic color set used by the TUI.
type Palette struct {
	Background string
	Foreground string
	Accent     string
	Correct    string
	Error      string
	Muted      string
	Colorless  bool
	Source     string
}

// DefaultPalette is ink-and-gold when no OS theme is available.
func DefaultPalette() Palette {
	return Palette{
		Background: "#140e0c",
		Foreground: "#e6d3a3",
		Accent:     "#c4a35a",
		Correct:    "#9bbf7a",
		Error:      "#b42318",
		Muted:      "#d7dde3",
		Source:     "ink-fallback",
	}
}

// CyberpunkPalette is kept as an alias so older tests and call sites still compile.
func CyberpunkPalette() Palette { return DefaultPalette() }

// LoadPalette resolves a theme in this order:
// NO_COLOR, BLACKLETTER_THEME, Omarchy current theme, ink fallback.
func LoadPalette() Palette {
	if os.Getenv("NO_COLOR") != "" {
		p := DefaultPalette()
		p.Colorless = true
		p.Source = "no-color"
		return p
	}
	if path := os.Getenv("BLACKLETTER_THEME"); path != "" {
		if p, err := LoadPaletteFile(path); err == nil {
			p.Source = path
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err == nil {
		for _, rel := range []string{
			".local/state/omarchy/current/theme/colors.toml",
			".config/omarchy/current/theme/colors.toml",
		} {
			path := filepath.Join(home, rel)
			if p, err := LoadPaletteFile(path); err == nil {
				p.Source = path
				return p
			}
		}
	}
	return DefaultPalette()
}

// LoadPaletteFile reads an Omarchy-style colors.toml from path.
func LoadPaletteFile(path string) (Palette, error) {
	f, err := os.Open(path)
	if err != nil {
		return Palette{}, err
	}
	defer f.Close()
	return PaletteFromTOML(f)
}

// PaletteFromTOML maps a simple TOML color map onto semantic roles.
func PaletteFromTOML(r io.Reader) (Palette, error) {
	keys, err := parseSimpleTOML(r)
	if err != nil {
		return Palette{}, err
	}
	if len(keys) == 0 {
		return DefaultPalette(), nil
	}
	p := DefaultPalette()
	p.Background = firstColor(keys, p.Background, "background")
	p.Foreground = firstColor(keys, p.Foreground, "bright_foreground", "foreground")
	p.Accent = firstColor(keys, p.Accent, "accent", "cyan", "color6")
	p.Correct = firstColor(keys, p.Correct, "green", "color2")
	p.Error = firstColor(keys, p.Error, "red", "color1")
	p.Muted = firstColor(keys, p.Muted, "muted", "dark_foreground")
	p.Source = "toml"
	return LiftMuted(p), nil
}

// LiftMuted replaces a near-black muted color with a readable silver.
// Theme files often set muted/dark_foreground too dark for terminal text.
func LiftMuted(p Palette) Palette {
	if p.Colorless {
		return p
	}
	if tooDark(p.Muted) {
		p.Muted = "#d7dde3"
	}
	return p
}

func tooDark(hex string) bool {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return true
	}
	// Rec. 601 luma, 0–255. Below ~140 disappears on a dark Mac terminal.
	return (int(r)*299+int(g)*587+int(b)*114)/1000 < 140
}

func parseHex(hex string) (r, g, b byte, ok bool) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	var n int
	for i := 0; i < 6; i++ {
		c := s[i]
		var v byte
		switch {
		case c >= '0' && c <= '9':
			v = c - '0'
		case c >= 'a' && c <= 'f':
			v = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			v = c - 'A' + 10
		default:
			return 0, 0, 0, false
		}
		n = n<<4 | int(v)
	}
	return byte(n >> 16), byte(n >> 8), byte(n), true
}

func firstColor(keys map[string]string, fallback string, names ...string) string {
	for _, n := range names {
		if v, ok := keys[n]; ok && v != "" {
			return v
		}
	}
	return fallback
}

func parseSimpleTOML(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = unquoteTOML(v)
		if k != "" && v != "" {
			out[k] = v
		}
	}
	return out, sc.Err()
}

func unquoteTOML(v string) string {
	if strings.HasPrefix(v, `"`) {
		if end := strings.Index(v[1:], `"`); end >= 0 {
			return v[1 : 1+end]
		}
	}
	if strings.HasPrefix(v, `'`) {
		if end := strings.Index(v[1:], `'`); end >= 0 {
			return v[1 : 1+end]
		}
	}
	if i := strings.Index(v, "#"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return v
}
