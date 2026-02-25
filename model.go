package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	btOnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	btOffStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5F56"))

	scanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Italic(true)

	noDevicesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Bold(true).
			MarginTop(1)
)

type BluetoothDevice struct {
	MAC        string
	Name       string
	Connected  bool
	Paired     bool
	Trusted    bool
	DeviceType string
}

type Model struct {
	devices          []BluetoothDevice
	cursor           int
	scanning         bool
	width            int
	height           int
	bluetoothEnabled bool
	bluetoothChecked bool
	statusText       string
}

type scanCompleteMsg struct {
	devices []BluetoothDevice
}

type deviceStatusMsg struct {
	deviceMAC string
	connected bool
}

type bluetoothStatusMsg struct {
	enabled bool
	error   error
}

type errorMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		getDevicesCmd(),
		getBluetoothStatusCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	case scanCompleteMsg:
		m.scanning = false
		m.devices = msg.devices
		m.statusText = ""
		m.clampCursor()

	case deviceStatusMsg:
		return m.handleDeviceStatusMsg(msg)

	case []BluetoothDevice:
		m.devices = msg
		m.statusText = ""
		m.clampCursor()

	case bluetoothStatusMsg:
		return m.handleBluetoothStatusMsg(msg)

	case errorMsg:
		m.statusText = msg.err.Error()
	}

	return m, nil
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.devices) {
		m.cursor = max(0, len(m.devices)-1)
	}
}

func (m Model) deviceListOffset() int {
	offset := 2 // title + blank line
	if m.bluetoothChecked {
		offset++ // bluetooth status line
	}
	if m.scanning {
		offset += 2 // scanning text + blank line
	}
	return offset
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.statusText = ""

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.devices)-1 {
			m.cursor++
		}

	case "enter", " ":
		return m.handleDeviceAction()

	case "s":
		if !m.scanning {
			m.scanning = true
			return m, scanDevicesCmd()
		}

	case "r":
		return m, getDevicesCmd()

	case "d":
		return m.handleDisconnectAction()

	case "p":
		return m.handlePairAction()

	case "e":
		return m.handleBluetoothToggle()

	case "ctrl+r":
		return m, tea.Batch(
			getDevicesCmd(),
			getBluetoothStatusCmd(),
		)
	}

	return m, nil
}

func (m Model) handleDeviceAction() (tea.Model, tea.Cmd) {
	if len(m.devices) == 0 {
		return m, nil
	}

	device := m.devices[m.cursor]
	if device.Connected {
		return m, disconnectDeviceCmd(device.MAC)
	} else if device.Paired {
		return m, connectDeviceCmd(device.MAC)
	} else {
		return m, pairAndConnectDeviceCmd(device.MAC)
	}
}

func (m Model) handleDisconnectAction() (tea.Model, tea.Cmd) {
	if len(m.devices) > 0 {
		device := m.devices[m.cursor]
		if device.Connected {
			return m, disconnectDeviceCmd(device.MAC)
		}
	}
	return m, nil
}

func (m Model) handlePairAction() (tea.Model, tea.Cmd) {
	if len(m.devices) > 0 {
		device := m.devices[m.cursor]
		if !device.Paired {
			return m, pairDeviceCmd(device.MAC)
		}
	}
	return m, nil
}

func (m Model) handleBluetoothToggle() (tea.Model, tea.Cmd) {
	if m.bluetoothChecked {
		if m.bluetoothEnabled {
			return m, disableBluetoothCmd()
		} else {
			return m, enableBluetoothCmd()
		}
	}
	return m, nil
}

func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.MouseButtonWheelDown:
		if m.cursor < len(m.devices)-1 {
			m.cursor++
		}
	case tea.MouseButtonLeft:
		offset := m.deviceListOffset()
		if msg.Y >= offset && msg.Y < offset+len(m.devices) {
			newCursor := msg.Y - offset
			if newCursor >= 0 && newCursor < len(m.devices) {
				m.cursor = newCursor
			}
		}
	}
	return m, nil
}

func (m Model) handleDeviceStatusMsg(msg deviceStatusMsg) (tea.Model, tea.Cmd) {
	for i, device := range m.devices {
		if device.MAC == msg.deviceMAC {
			m.devices[i].Connected = msg.connected
			break
		}
	}
	return m, getDevicesCmd()
}

func (m Model) handleBluetoothStatusMsg(msg bluetoothStatusMsg) (tea.Model, tea.Cmd) {
	m.bluetoothEnabled = msg.enabled
	m.bluetoothChecked = true
	if msg.error == nil {
		m.statusText = ""
		return m, getDevicesCmd()
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("HyprBluetooth - Bluetooth Device Manager"))
	s.WriteString("\n")

	if m.bluetoothChecked {
		if m.bluetoothEnabled {
			s.WriteString(btOnStyle.Render("🔵 Bluetooth: ON"))
		} else {
			s.WriteString(btOffStyle.Render("🔴 Bluetooth: OFF"))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")

	if m.scanning {
		s.WriteString(scanStyle.Render("🔍 Scanning for devices..."))
		s.WriteString("\n\n")
	}

	if !m.bluetoothEnabled && m.bluetoothChecked {
		s.WriteString(disabledStyle.Render("Bluetooth is disabled. Press 'e' to enable."))
		s.WriteString("\n")
	} else if len(m.devices) == 0 {
		s.WriteString(noDevicesStyle.Render("No devices found. Press 's' to scan for devices."))
		s.WriteString("\n")
	} else {
		for i, device := range m.devices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			status := "○"
			statusColor := "#626262"
			if device.Connected {
				status = "●"
				statusColor = "#04B575"
			} else if device.Paired {
				status = "◐"
				statusColor = "#FFA500"
			}

			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))

			deviceName := device.Name
			if deviceName == "" {
				deviceName = "Unknown Device"
			}

			itemStyle := lipgloss.NewStyle()
			if m.cursor == i {
				itemStyle = itemStyle.Background(lipgloss.Color("#383838"))
			}

			line := fmt.Sprintf("%s %s %s (%s)",
				cursor,
				statusStyle.Render(status),
				deviceName,
				device.MAC)

			s.WriteString(itemStyle.Render(line))
			s.WriteString("\n")
		}
	}

	if m.statusText != "" {
		s.WriteString(errorStyle.Render("Error: " + m.statusText))
		s.WriteString("\n")
	}

	help := `
Controls:
  ↑/k, ↓/j: Navigate  Enter/Space: Connect/Disconnect  s: Scan  r: Refresh
  p: Pair  d: Disconnect  e: Enable/Disable Bluetooth  Ctrl+r: Full Refresh  q: Quit

Status: ● Connected  ◐ Paired  ○ Unpaired`

	s.WriteString(helpStyle.Render(help))

	return s.String()
}
