package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Dandarprox/tuiference/internal/tui"
)

func main() {
	program := tea.NewProgram(tui.New(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tuiference: %v\n", err)
		os.Exit(1)
	}
}
