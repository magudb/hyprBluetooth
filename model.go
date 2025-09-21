package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	selected         map[int]struct{}
	width            int
	height           int
	bluetoothEnabled bool
	bluetoothChecked bool
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

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		getDevicesCmd(),
		getBluetoothStatusCmd(),
		tea.EnterAltScreen,
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

	case deviceStatusMsg:
		return m.handleDeviceStatusMsg(msg)

	case []BluetoothDevice:
		m.devices = msg

	case bluetoothStatusMsg:
		return m.handleBluetoothStatusMsg(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonWheelUp {
			if m.cursor > 0 {
				m.cursor--
			}
		} else if msg.Button == tea.MouseButtonWheelDown {
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			}
		} else if msg.Button == tea.MouseButtonLeft {
			if msg.Y >= 3 && msg.Y < 3+len(m.devices) {
				newCursor := msg.Y - 3
				if newCursor >= 0 && newCursor < len(m.devices) {
					m.cursor = newCursor
				}
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
		return m, getDevicesCmd()
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Bold(true)

	s.WriteString(titleStyle.Render("HyprBluetooth - Bluetooth Device Manager"))
	s.WriteString("\n")

	if m.bluetoothChecked {
		statusStyle := lipgloss.NewStyle().Bold(true)
		if m.bluetoothEnabled {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#04B575"))
			s.WriteString(statusStyle.Render("🔵 Bluetooth: ON"))
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#FF5F56"))
			s.WriteString(statusStyle.Render("🔴 Bluetooth: OFF"))
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")

	if m.scanning {
		scanStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
		s.WriteString(scanStyle.Render("🔍 Scanning for devices..."))
		s.WriteString("\n\n")
	}

	if !m.bluetoothEnabled && m.bluetoothChecked {
		disabledStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Italic(true)
		s.WriteString(disabledStyle.Render("Bluetooth is disabled. Press 'e' to enable."))
		s.WriteString("\n")
	} else if len(m.devices) == 0 {
		noDevicesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
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

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginTop(2)

	help := `
Controls:
  ↑/k, ↓/j: Navigate  Enter/Space: Connect/Disconnect  s: Scan  r: Refresh
  p: Pair  d: Disconnect  e: Enable/Disable Bluetooth  Ctrl+r: Full Refresh  q: Quit

Status: ● Connected  ◐ Paired  ○ Unpaired`

	s.WriteString(helpStyle.Render(help))

	return s.String()
}
