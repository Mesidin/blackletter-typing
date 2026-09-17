package app

import (
	"strings"
	"testing"
)

func TestFingerFor(t *testing.T) {
	if FingerFor('f') != "left index" {
		t.Fatalf("f %q", FingerFor('f'))
	}
	if FingerFor('j') != "right index" {
		t.Fatalf("j %q", FingerFor('j'))
	}
	if FingerFor(' ') != "thumbs" {
		t.Fatalf("space %q", FingerFor(' '))
	}
	if FingerFor('A') != "left pinky + shift" {
		t.Fatalf("A %q", FingerFor('A'))
	}
}

func TestRenderKeyboardLooksLikeKeys(t *testing.T) {
	st := NewStyles(Palette{Colorless: true})
	out := RenderKeyboard(st, 'f', 80, 24)
	if !strings.Contains(out, "F") {
		t.Fatalf("missing F cap:\n%s", out)
	}
	if !strings.Contains(out, "SPACE") {
		t.Fatalf("missing space bar:\n%s", out)
	}
	if !strings.ContainsAny(out, "─━═") && !strings.ContainsAny(out, "┌╭╔") {
		t.Fatalf("expected bordered keycaps:\n%s", out)
	}
	if strings.Count(out, "Q") < 1 || strings.Count(out, "A") < 1 || strings.Count(out, "Z") < 1 {
		t.Fatalf("missing rows:\n%s", out)
	}
	t.Log("\n" + out)
}

func TestKeyLabel(t *testing.T) {
	if keyLabel('f') != "F" {
		t.Fatalf("f -> %q", keyLabel('f'))
	}
	if keyLabel(' ') != "SPACE" {
		t.Fatalf("space -> %q", keyLabel(' '))
	}
	if keyLabel(';') != ";" {
		t.Fatalf("; -> %q", keyLabel(';'))
	}
}

func TestKeyPaddingGrows(t *testing.T) {
	x1, y1 := keyPadding(60, 20)
	x2, _ := keyPadding(80, 24)
	x3, y3 := keyPadding(110, 40)
	if x2 < x1 {
		t.Fatalf("wider terminal should not shrink padX: %d vs %d", x2, x1)
	}
	if x3 <= x2 {
		t.Fatalf("very wide terminal should grow padX: %d vs %d", x3, x2)
	}
	if y3 <= y1 {
		t.Fatalf("tall terminal should add vertical pad: %d vs %d", y3, y1)
	}
}
