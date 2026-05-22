package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	bluetoothYes         = "yes"
	cmdTimeout           = 15 * time.Second
	scanCmdTimeout       = 30 * time.Second
	scanDuration         = 5 * time.Second
	postPairConnectDelay = 1 * time.Second
	infoFetchConcurrency = 4
)

var macRegex = regexp.MustCompile(`^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}$`)

// runBluetoothctl and runBluetoothctlCombined are overridable to enable testing.
var runBluetoothctl = func(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "bluetoothctl", args...).Output()
}

var runBluetoothctlCombined = func(ctx context.Context, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, "bluetoothctl", args...).CombinedOutput()
}

type BluetoothDevice struct {
	MAC       string
	Name      string
	Connected bool
	Paired    bool
	Trusted   bool
}

func validateMAC(mac string) error {
	if !macRegex.MatchString(mac) {
		return fmt.Errorf("invalid MAC address: %q", mac)
	}
	return nil
}

func parseDevicesOutput(b []byte) []BluetoothDevice {
	var out []BluetoothDevice
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 || parts[0] != "Device" {
			continue
		}
		mac := parts[1]
		if !macRegex.MatchString(mac) {
			continue
		}
		name := ""
		if len(parts) > 2 {
			name = parts[2]
		}
		out = append(out, BluetoothDevice{MAC: mac, Name: name})
	}
	return out
}

func parseDeviceInfo(b []byte, mac string) BluetoothDevice {
	d := BluetoothDevice{MAC: mac}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case strings.HasPrefix(line, "Name: "):
			d.Name = strings.TrimPrefix(line, "Name: ")
		case strings.HasPrefix(line, "Connected: "):
			d.Connected = strings.TrimPrefix(line, "Connected: ") == bluetoothYes
		case strings.HasPrefix(line, "Paired: "):
			d.Paired = strings.TrimPrefix(line, "Paired: ") == bluetoothYes
		case strings.HasPrefix(line, "Trusted: "):
			d.Trusted = strings.TrimPrefix(line, "Trusted: ") == bluetoothYes
		}
	}
	return d
}

func parsePoweredStatus(b []byte) (bool, error) {
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if v, ok := strings.CutPrefix(line, "Powered: "); ok {
			return v == bluetoothYes, nil
		}
	}
	return false, errors.New("could not determine bluetooth status")
}

func getDeviceInfo(ctx context.Context, mac string) (BluetoothDevice, error) {
	if err := validateMAC(mac); err != nil {
		return BluetoothDevice{}, err
	}
	output, err := runBluetoothctl(ctx, "info", mac)
	if err != nil {
		return BluetoothDevice{}, fmt.Errorf("failed to get device info: %w", err)
	}
	return parseDeviceInfo(output, mac), nil
}

func getDevices(ctx context.Context) ([]BluetoothDevice, error) {
	output, err := runBluetoothctl(ctx, "devices")
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	devices := parseDevicesOutput(output)

	sem := make(chan struct{}, infoFetchConcurrency)
	var wg sync.WaitGroup
	for i := range devices {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			info, err := getDeviceInfo(ctx, devices[i].MAC)
			if err != nil {
				return
			}
			if info.Name != "" {
				devices[i].Name = info.Name
			}
			devices[i].Connected = info.Connected
			devices[i].Paired = info.Paired
			devices[i].Trusted = info.Trusted
		}(i)
	}
	wg.Wait()
	return devices, nil
}

// scanDevices toggles discovery on, waits, then toggles it off. The "off"
// is always attempted even if the wait is canceled.
func scanDevices(ctx context.Context) ([]BluetoothDevice, error) {
	startCtx, cancelStart := context.WithTimeout(ctx, cmdTimeout)
	defer cancelStart()
	if output, err := runBluetoothctlCombined(startCtx, "scan", "on"); err != nil {
		return nil, fmt.Errorf("failed to start scan: %w, output: %s", err, string(output))
	}

	defer func() {
		// Use Background so we still stop scanning if the outer context is canceled.
		stopCtx, cancelStop := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancelStop()
		_, _ = runBluetoothctlCombined(stopCtx, "scan", "off")
	}()

	select {
	case <-time.After(scanDuration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return getDevices(ctx)
}

func connectDevice(ctx context.Context, mac string) error {
	if err := validateMAC(mac); err != nil {
		return err
	}
	output, err := runBluetoothctlCombined(ctx, "connect", mac)
	if err != nil {
		return fmt.Errorf("failed to connect to device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func disconnectDevice(ctx context.Context, mac string) error {
	if err := validateMAC(mac); err != nil {
		return err
	}
	output, err := runBluetoothctlCombined(ctx, "disconnect", mac)
	if err != nil {
		return fmt.Errorf("failed to disconnect from device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func pairDevice(ctx context.Context, mac string) error {
	if err := validateMAC(mac); err != nil {
		return err
	}
	output, err := runBluetoothctlCombined(ctx, "pair", mac)
	if err != nil {
		return fmt.Errorf("failed to pair with device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func trustDevice(ctx context.Context, mac string) error {
	if err := validateMAC(mac); err != nil {
		return err
	}
	output, err := runBluetoothctlCombined(ctx, "trust", mac)
	if err != nil {
		return fmt.Errorf("failed to trust device %s: %w, output: %s", mac, err, string(output))
	}
	return nil
}

func isBluetoothEnabled(ctx context.Context) (bool, error) {
	output, err := runBluetoothctl(ctx, "show")
	if err != nil {
		return false, fmt.Errorf("failed to get bluetooth status: %w", err)
	}
	return parsePoweredStatus(output)
}

func enableBluetooth(ctx context.Context) error {
	output, err := runBluetoothctlCombined(ctx, "power", "on")
	if err != nil {
		return fmt.Errorf("failed to enable bluetooth: %w, output: %s", err, string(output))
	}
	return nil
}

func disableBluetooth(ctx context.Context) error {
	output, err := runBluetoothctlCombined(ctx, "power", "off")
	if err != nil {
		return fmt.Errorf("failed to disable bluetooth: %w, output: %s", err, string(output))
	}
	return nil
}

// Bubble Tea command factories

func getDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		devices, err := getDevices(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return devicesMsg{devices: devices}
	}
}

func scanDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), scanCmdTimeout)
		defer cancel()
		devices, err := scanDevices(ctx)
		if err != nil {
			return scanCompleteMsg{err: err}
		}
		return scanCompleteMsg{devices: devices}
	}
}

func connectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := connectDevice(ctx, mac); err != nil {
			return errorMsg{err: err}
		}
		return deviceStatusMsg{deviceMAC: mac, connected: true}
	}
}

func disconnectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := disconnectDevice(ctx, mac); err != nil {
			return errorMsg{err: err}
		}
		return deviceStatusMsg{deviceMAC: mac, connected: false}
	}
}

func pairDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := pairDevice(ctx, mac); err != nil {
			return errorMsg{err: err}
		}
		if err := trustDevice(ctx, mac); err != nil {
			return errorMsg{err: fmt.Errorf("paired but failed to trust: %w", err)}
		}
		devices, err := getDevices(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return devicesMsg{devices: devices}
	}
}

func pairAndConnectDeviceCmd(mac string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), scanCmdTimeout)
		defer cancel()
		if err := pairDevice(ctx, mac); err != nil {
			return errorMsg{err: err}
		}
		if err := trustDevice(ctx, mac); err != nil {
			return errorMsg{err: fmt.Errorf("paired but failed to trust: %w", err)}
		}
		select {
		case <-time.After(postPairConnectDelay):
		case <-ctx.Done():
			return errorMsg{err: ctx.Err()}
		}
		if err := connectDevice(ctx, mac); err != nil {
			return errorMsg{err: fmt.Errorf("paired but failed to connect: %w", err)}
		}
		devices, err := getDevices(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return devicesMsg{devices: devices}
	}
}

func getBluetoothStatusCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		enabled, err := isBluetoothEnabled(ctx)
		if err != nil {
			return bluetoothStatusMsg{enabled: false, err: err}
		}
		return bluetoothStatusMsg{enabled: enabled}
	}
}

func enableBluetoothCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := enableBluetooth(ctx); err != nil {
			return errorMsg{err: err}
		}
		return bluetoothStatusMsg{enabled: true}
	}
}

func disableBluetoothCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := disableBluetooth(ctx); err != nil {
			return errorMsg{err: err}
		}
		return bluetoothStatusMsg{enabled: false}
	}
}
