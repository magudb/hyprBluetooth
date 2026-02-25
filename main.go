package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"     //nolint:unused // set via ldflags at build time
	commit  = "none"    //nolint:unused // set via ldflags at build time
	date    = "unknown" //nolint:unused // set via ldflags at build time
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() Model {
	return Model{
		devices:          []BluetoothDevice{},
		cursor:           0,
		scanning:         false,
		width:            80,
		height:           24,
		bluetoothEnabled: false,
		bluetoothChecked: false,
	}
}
