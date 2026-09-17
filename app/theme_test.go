package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaletteFromTOML(t *testing.T) {
	src := `
mode = "dark"
accent = "#7aa2f7"
background = "#1a1b26"
foreground = "#a9b1d6"
bright_foreground = "#c0caf5"
muted = "#414868"
green = "#9ece6a"
red = "#f7768e"
color2 = "#9ece6a"
color1 = "#f7768e"
# comment
cyan = "#7dcfff"
`
	p, err := PaletteFromTOML(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if p.Background != "#1a1b26" {
		t.Fatalf("background %q", p.Background)
	}
	if p.Foreground != "#c0caf5" {
		t.Fatalf("foreground prefers bright_foreground, got %q", p.Foreground)
	}
	if p.Accent != "#7aa2f7" {
		t.Fatalf("accent %q", p.Accent)
	}
	if p.Correct != "#9ece6a" {
		t.Fatalf("correct %q", p.Correct)
	}
	if p.Error != "#f7768e" {
		t.Fatalf("error %q", p.Error)
	}
	if p.Muted != "#d7dde3" {
		t.Fatalf("dark theme muted should be lifted for contrast, got %q", p.Muted)
	}
}

func TestPaletteFromTOMLInlineComment(t *testing.T) {
	src := `accent = "#7aa2f7"  # Main accent`
	p, err := PaletteFromTOML(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if p.Accent != "#7aa2f7" {
		t.Fatalf("accent %q", p.Accent)
	}
}

func TestLoadPaletteNOCOLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("BLACKLETTER_THEME", "")
	p := LoadPalette()
	if !p.Colorless {
		t.Fatal("expected colorless")
	}
}

func TestLoadPaletteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "colors.toml")
	if err := os.WriteFile(path, []byte("accent = \"#ff00aa\"\nbackground = \"#000000\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPaletteFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.Accent != "#ff00aa" {
		t.Fatalf("accent %q", p.Accent)
	}
}

func TestCyberpunkFallback(t *testing.T) {
	p := DefaultPalette()
	if p.Correct != "#9bbf7a" || p.Error != "#b42318" || p.Accent != "#c4a35a" {
		t.Fatalf("unexpected fallback %+v", p)
	}
	if tooDark(p.Muted) {
		t.Fatalf("fallback muted too dark for a Mac terminal: %q", p.Muted)
	}
}

func TestLiftMuted(t *testing.T) {
	p := Palette{Muted: "#333333", Colorless: false}
	got := LiftMuted(p)
	if tooDark(got.Muted) {
		t.Fatalf("still too dark: %q", got.Muted)
	}
	bright := Palette{Muted: "#c0caf5"}
	if LiftMuted(bright).Muted != "#c0caf5" {
		t.Fatal("should keep a light muted color")
	}
}

func TestTooDark(t *testing.T) {
	if !tooDark("#333333") {
		t.Fatal("#333333 should be too dark")
	}
	if !tooDark("#414868") {
		t.Fatal("tokyo-night muted should be too dark")
	}
	if tooDark("#d7dde3") {
		t.Fatal("silver should pass")
	}
}
