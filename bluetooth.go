package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const bluetoothYes = "yes"

type BluetoothManager struct{}

func NewBluetoothManager() *BluetoothManager {
	return &BluetoothManager{}
}

func (bm *BluetoothManager) GetDevices() ([]BluetoothDevice, error) {
	cmd := exec.Command("bluetoothctl", "devices")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	devices := make([]BluetoothDevice, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 || parts[0] != "Device" {
			continue
		}

		mac := parts[1]
		name := ""
		if len(parts) > 2 {
			name = parts[2]
		}

		device := BluetoothDevice{
			MAC:  mac,
			Name: name,
		}

		info, err := bm.GetDeviceInfo(mac)
		if err == nil {
			device.Connected = info.Connected
			device.Paired = info.Paired
			device.Trusted = info.Trusted
			device.DeviceType = info.DeviceType
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func (bm *BluetoothManager) GetDeviceInfo(mac string) (*BluetoothDevice, error) {
	cmd := exec.Command("bluetoothctl", "info", mac)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	device := &BluetoothDevice{MAC: mac}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Name: ") {
			device.Name = strings.TrimPrefix(line, "Name: ")
		} else if strings.HasPrefix(line, "Connected: ") {
			device.Connected = strings.TrimPrefix(line, "Connected: ") == bluetoothYes
		} else if strings.HasPrefix(line, "Paired: ") {
			device.Paired = strings.TrimPrefix(line, "Paired: ") == bluetoothYes
		} else if strings.HasPrefix(line, "Trusted: ") {
			device.Trusted = strings.TrimPrefix(line, "Trusted: ") == bluetoothYes
		} else if strings.HasPrefix(line, "Icon: ") {
			device.DeviceType = strings.TrimPrefix(line, "Icon: ")
		}
	}

	return device, nil
}

func (bm *BluetoothManager) ScanDevices() ([]BluetoothDevice, error) {
	startCmd := exec.Command("bluetoothctl", "scan", "on")
	if err := startCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to start scan: %w", err)
	}

	time.Sleep(5 * time.Second)

	stopCmd := exec.Command("bluetoothctl", "scan", "off")
	if err := stopCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to stop scan: %w", err)
	}

	return bm.GetDevices()
}

func (bm *BluetoothManager) ConnectDevice(mac string) error {
	cmd := exec.Command("bluetoothctl", "connect", mac)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect to device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func (bm *BluetoothManager) DisconnectDevice(mac string) error {
	cmd := exec.Command("bluetoothctl", "disconnect", mac)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disconnect from device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func (bm *BluetoothManager) PairDevice(mac string) error {
	cmd := exec.Command("bluetoothctl", "pair", mac)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pair with device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func (bm *BluetoothManager) TrustDevice(mac string) error {
	cmd := exec.Command("bluetoothctl", "trust", mac)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to trust device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func (bm *BluetoothManager) IsBluetoothEnabled() (bool, error) {
	cmd := exec.Command("bluetoothctl", "show")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get bluetooth status: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Powered: ") {
			return strings.TrimPrefix(line, "Powered: ") == "yes", nil
		}
	}

	return false, fmt.Errorf("could not determine bluetooth status")
}

func (bm *BluetoothManager) EnableBluetooth() error {
	cmd := exec.Command("bluetoothctl", "power", "on")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable bluetooth: %w, output: %s", err, string(output))
	}
	return nil
}

func (bm *BluetoothManager) DisableBluetooth() error {
	cmd := exec.Command("bluetoothctl", "power", "off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disable bluetooth: %w, output: %s", err, string(output))
	}
	return nil
}

// Bubble Tea command functions

func getDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		devices, err := bm.GetDevices()
		if err != nil {
			return errorMsg{err: err}
		}
		return devices
	}
}

func scanDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		devices, err := bm.ScanDevices()
		if err != nil {
			return errorMsg{err: err}
		}
		return scanCompleteMsg{devices: devices}
	}
}

func connectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.ConnectDevice(mac); err != nil {
			return errorMsg{err: err}
		}
		return deviceStatusMsg{
			deviceMAC: mac,
			connected: true,
		}
	}
}

func disconnectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.DisconnectDevice(mac); err != nil {
			return errorMsg{err: err}
		}
		return deviceStatusMsg{
			deviceMAC: mac,
			connected: false,
		}
	}
}

func pairDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.PairDevice(mac); err != nil {
			return errorMsg{err: err}
		}
		_ = bm.TrustDevice(mac)
		return getDevicesCmd()()
	}
}

func pairAndConnectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.PairDevice(mac); err != nil {
			return errorMsg{err: err}
		}
		_ = bm.TrustDevice(mac)
		time.Sleep(1 * time.Second)
		_ = bm.ConnectDevice(mac)
		return getDevicesCmd()()
	}
}

func getBluetoothStatusCmd() tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		enabled, err := bm.IsBluetoothEnabled()
		if err != nil {
			return bluetoothStatusMsg{enabled: false, error: err}
		}
		return bluetoothStatusMsg{enabled: enabled, error: nil}
	}
}

func enableBluetoothCmd() tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.EnableBluetooth(); err != nil {
			return errorMsg{err: err}
		}
		return bluetoothStatusMsg{enabled: true, error: nil}
	}
}

func disableBluetoothCmd() tea.Cmd {
	return func() tea.Msg {
		bm := NewBluetoothManager()
		if err := bm.DisableBluetooth(); err != nil {
			return errorMsg{err: err}
		}
		return bluetoothStatusMsg{enabled: false, error: nil}
	}
}
