package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("hyprBluetooth %s (commit %s, built %s)\n", version, commit, date)
			return
		case "--help", "-h", "help":
			printUsage()
			return
		}
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`hyprBluetooth - terminal Bluetooth device manager

Usage:
  hyprBluetooth          launch the interactive TUI
  hyprBluetooth --help   show this message
  hyprBluetooth --version  print version information`)
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
