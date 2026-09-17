package app

import "strings"

// Fraktur maps Latin letters to Mathematical Bold Fraktur.
// That Unicode block is the closest thing a terminal has to a blackletter face.
func Fraktur(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 4)
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(0x1D56C + (r - 'A'))
		case r >= 'a' && r <= 'z':
			b.WriteRune(0x1D586 + (r - 'a'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FrakturBanner is a one-line gothic title with simple ornaments.
func FrakturBanner(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return "✠  " + Fraktur(s) + "  ✠"
}

func appWordmark() string {
	return "✠  " + Fraktur("Blackletter") + "  ✠"
}
