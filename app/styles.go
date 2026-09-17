package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	minWidth  = 60
	minHeight = 18
)

// Styles holds lipgloss styles derived from a Palette.
type Styles struct {
	Palette Palette

	Title            lipgloss.Style
	Subtitle         lipgloss.Style
	Box              lipgloss.Style
	Help             lipgloss.Style
	Correct          lipgloss.Style
	Wrong            lipgloss.Style
	Caret            lipgloss.Style
	Muted            lipgloss.Style
	Accent           lipgloss.Style
	Subtle           lipgloss.Style
	Success          lipgloss.Style
	Error            lipgloss.Style
	MenuSelected     lipgloss.Style
	MenuSelectedDesc lipgloss.Style
	MenuNormal       lipgloss.Style
	MenuNormalDesc   lipgloss.Style
	Label            lipgloss.Style
	Value            lipgloss.Style
	Key              lipgloss.Style
	KeyActive        lipgloss.Style
}

// NewStyles builds the cyberpunk (or colorless) style set.
func NewStyles(p Palette) Styles {
	if p.Colorless {
		return colorlessStyles(p)
	}
	p = LiftMuted(p)
	fg := lipgloss.Color(p.Foreground)
	accent := lipgloss.Color(p.Accent)
	ok := lipgloss.Color(p.Correct)
	bad := lipgloss.Color(p.Error)
	// Adaptive so chrome stays readable on both dark Mac terminals and light ones.
	// Do not use the theme's near-black "muted" for anything a kid has to read.
	secondary := lipgloss.AdaptiveColor{Light: "#1f4d32", Dark: "#b6f5d0"}
	unread := lipgloss.AdaptiveColor{Light: "#3d4550", Dark: "#d7dde3"}

	s := Styles{Palette: p}
	s.Title = lipgloss.NewStyle().Foreground(accent).Bold(true)
	s.Subtitle = lipgloss.NewStyle().Foreground(fg)
	s.Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1)
	s.Help = lipgloss.NewStyle().Foreground(secondary)
	s.Correct = lipgloss.NewStyle().Foreground(ok).Bold(true)
	s.Wrong = lipgloss.NewStyle().Foreground(bad).Bold(true)
	s.Caret = lipgloss.NewStyle().Foreground(accent).Bold(true).Underline(true)
	s.Muted = lipgloss.NewStyle().Foreground(unread)
	s.Accent = lipgloss.NewStyle().Foreground(accent).Bold(true)
	s.Subtle = lipgloss.NewStyle().Foreground(secondary)
	s.Success = lipgloss.NewStyle().Foreground(ok).Bold(true)
	s.Error = lipgloss.NewStyle().Foreground(bad).Bold(true)
	s.MenuSelected = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(accent).
		Foreground(accent).
		Bold(true).
		Padding(0, 0, 0, 1)
	s.MenuSelectedDesc = s.MenuSelected.Bold(false).Foreground(fg)
	s.MenuNormal = lipgloss.NewStyle().Foreground(fg).Padding(0, 0, 0, 2)
	s.MenuNormalDesc = lipgloss.NewStyle().Foreground(secondary).Padding(0, 0, 0, 2)
	s.Label = lipgloss.NewStyle().Foreground(secondary)
	s.Value = lipgloss.NewStyle().Foreground(fg).Bold(true)
	s.Key = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(unread).
		Foreground(fg).
		Bold(true)
	ink := lipgloss.Color(p.Background)
	if p.Background == "" {
		ink = lipgloss.Color("#0d0208")
	}
	s.KeyActive = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(accent).
		Foreground(ink).
		Background(accent).
		Bold(true)
	return s
}

func colorlessStyles(p Palette) Styles {
	s := Styles{Palette: p}
	s.Title = lipgloss.NewStyle().Bold(true)
	s.Subtitle = lipgloss.NewStyle()
	s.Box = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	s.Help = lipgloss.NewStyle()
	s.Correct = lipgloss.NewStyle().Bold(true)
	s.Wrong = lipgloss.NewStyle().Bold(true).Underline(true)
	s.Caret = lipgloss.NewStyle().Bold(true).Underline(true)
	s.Muted = lipgloss.NewStyle()
	s.Accent = lipgloss.NewStyle().Bold(true)
	s.Subtle = lipgloss.NewStyle()
	s.Success = lipgloss.NewStyle().Bold(true)
	s.Error = lipgloss.NewStyle().Bold(true)
	s.MenuSelected = lipgloss.NewStyle().Bold(true).Padding(0, 0, 0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true)
	s.MenuSelectedDesc = s.MenuSelected.Bold(false)
	s.MenuNormal = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	s.MenuNormalDesc = lipgloss.NewStyle().Padding(0, 0, 0, 2)
	s.Label = lipgloss.NewStyle()
	s.Value = lipgloss.NewStyle().Bold(true)
	s.Key = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Bold(true)
	s.KeyActive = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		Bold(true).
		Reverse(true)
	return s
}

// Wordmark is the ASCII title, already styled.
func (s Styles) Wordmark() string {
	return s.Title.Render(appWordmark())
}

func (s Styles) frame(width int, body string) string {
	if width < 20 {
		width = 20
	}
	return s.Box.Width(width - 2).Render(body)
}

func joinBlocks(parts ...string) string {
	var nonempty []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonempty = append(nonempty, p)
		}
	}
	return strings.Join(nonempty, "\n")
}
