package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func initialModel() Model {
	return Model{
		devices:          []BluetoothDevice{},
		cursor:           0,
		scanning:         false,
		selected:         make(map[int]struct{}),
		width:            80,
		height:           24,
		bluetoothEnabled: false,
		bluetoothChecked: false,
	}
}
