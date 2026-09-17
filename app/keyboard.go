package app

import (
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var kbRows = []string{
	"qwertyuiop",
	"asdfghjkl;",
	"zxcvbnm,./",
}

// FingerFor returns a kid-friendly finger label for a QWERTY key.
func FingerFor(r rune) string {
	shifted := unicode.IsUpper(r)
	base := fingerBase(unicode.ToLower(r))
	if shifted && base != "" && base != "thumbs" {
		return base + " + shift"
	}
	return base
}

func fingerBase(r rune) string {
	switch r {
	case '`', '1', 'q', 'a', 'z':
		return "left pinky"
	case '2', 'w', 's', 'x':
		return "left ring"
	case '3', 'e', 'd', 'c':
		return "left middle"
	case '4', '5', 'r', 't', 'f', 'g', 'v', 'b':
		return "left index"
	case '6', '7', 'y', 'u', 'h', 'j', 'n', 'm':
		return "right index"
	case '8', 'i', 'k', ',':
		return "right middle"
	case '9', 'o', 'l', '.':
		return "right ring"
	case '0', '-', '=', 'p', '[', ']', '\\', ';', '\'', '/':
		return "right pinky"
	case ' ':
		return "thumbs"
	default:
		return ""
	}
}

func keyPadding(maxWidth, maxHeight int) (padX, padY int) {
	padX = 1
	if maxWidth >= 78 {
		padX = 2
	}
	if maxWidth >= 100 {
		padX = 3
	}
	if maxHeight >= 32 {
		padY = 1
	}
	return padX, padY
}

func keyLabel(r rune) string {
	if r == ' ' {
		return "SPACE"
	}
	if r >= 'a' && r <= 'z' {
		return string(unicode.ToUpper(r))
	}
	return string(r)
}

func (s Styles) renderKey(label string, active bool, padX, padY, minInner int) string {
	st := s.Key
	if active {
		st = s.KeyActive
	}
	st = st.Padding(padY, padX).Align(lipgloss.Center, lipgloss.Center)
	if minInner > 1 {
		st = st.Width(minInner)
	}
	return st.Render(label)
}

// RenderKeyboard draws a QWERTY board of bordered keycaps with the next key lit.
func RenderKeyboard(st Styles, next rune, maxWidth, maxHeight int) string {
	next = unicode.ToLower(next)
	padX, padY := keyPadding(maxWidth, maxHeight)
	sample := st.renderKey("Q", false, padX, padY, 1)
	unit := lipgloss.Width(sample)
	if unit < 5 {
		unit = 5
	}

	var rows []string
	indents := []int{0, max(2, unit/3), max(4, unit/2)}
	for i, row := range kbRows {
		caps := make([]string, 0, len(row))
		for _, r := range row {
			caps = append(caps, st.renderKey(keyLabel(r), r == next, padX, padY, 1))
		}
		line := lipgloss.JoinHorizontal(lipgloss.Top, caps...)
		if ind := indents[i]; ind > 0 {
			line = lipgloss.NewStyle().MarginLeft(ind).Render(line)
		}
		rows = append(rows, line)
	}

	spaceInner := unit*5 - 2
	if spaceInner < 7 {
		spaceInner = 7
	}
	space := st.renderKey("SPACE", next == ' ', padX, padY, spaceInner)
	spaceIndent := max(6, unit*2)
	rows = append(rows, lipgloss.NewStyle().MarginLeft(spaceIndent).Render(space))
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
