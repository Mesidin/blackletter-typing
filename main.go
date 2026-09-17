package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"blackletter/app"
)

//go:embed assets/texts
var textsFS embed.FS

func main() {
	passages, err := app.LoadCorpus(textsFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "blackletter: load texts: %v\n", err)
		os.Exit(1)
	}
	store, err := app.OpenStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "blackletter: profiles: %v\n", err)
		os.Exit(1)
	}
	styles := app.NewStyles(app.LoadPalette())
	m := app.New(store, passages, styles)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "blackletter: %v\n", err)
		os.Exit(1)
	}
}
