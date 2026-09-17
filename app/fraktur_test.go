package app

import "testing"

func TestFrakturLetters(t *testing.T) {
	got := Fraktur("Dracula")
	if got == "Dracula" {
		t.Fatal("expected fraktur letters, got plain ASCII")
	}
	if Fraktur("Dracula") != Fraktur("Dracula") {
		t.Fatal("not stable")
	}
	plain := Fraktur("123 !")
	if plain != "123 !" {
		t.Fatalf("non-letters should pass through, got %q", plain)
	}
}

func TestFrakturBanner(t *testing.T) {
	b := FrakturBanner("The Wolfman")
	if b == "" || b == "The Wolfman" {
		t.Fatalf("banner %q", b)
	}
	if !bannerFits(b, 40) {
		t.Fatalf("one-line banner should fit a typical panel: %q", b)
	}
}
