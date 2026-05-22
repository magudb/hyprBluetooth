package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const (
	testMACHeadphones = "AA:BB:CC:DD:EE:FF"
	testMACMouse      = "11:22:33:44:55:66"
)

func TestParseDevicesOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []BluetoothDevice
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "single device with name",
			input: "Device AA:BB:CC:DD:EE:FF My Device\n",
			want: []BluetoothDevice{
				{MAC: testMACHeadphones, Name: "My Device"},
			},
		},
		{
			name: "multiple devices",
			input: "Device AA:BB:CC:DD:EE:FF Device One\n" +
				"Device 11:22:33:44:55:66 Device Two\n",
			want: []BluetoothDevice{
				{MAC: testMACHeadphones, Name: "Device One"},
				{MAC: testMACMouse, Name: "Device Two"},
			},
		},
		{
			name:  "device without name",
			input: "Device AA:BB:CC:DD:EE:FF\n",
			want: []BluetoothDevice{
				{MAC: testMACHeadphones, Name: ""},
			},
		},
		{
			name:  "invalid MAC is filtered",
			input: "Device NOT-A-MAC Some Name\n",
			want:  nil,
		},
		{
			name: "garbage lines ignored",
			input: "Header line\n" +
				"\n" +
				"Device AA:BB:CC:DD:EE:FF Good\n" +
				"Random text\n" +
				"  Device 11:22:33:44:55:66 Indented\n",
			want: []BluetoothDevice{
				{MAC: testMACHeadphones, Name: "Good"},
				{MAC: testMACMouse, Name: "Indented"},
			},
		},
		{
			name:  "name with spaces is preserved",
			input: "Device AA:BB:CC:DD:EE:FF Brand New Headphones X-1000\n",
			want: []BluetoothDevice{
				{MAC: testMACHeadphones, Name: "Brand New Headphones X-1000"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseDevicesOutput([]byte(tc.input))
			if len(got) != len(tc.want) {
				t.Fatalf("got %d devices, want %d: %+v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("device %d: got %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseDeviceInfo(t *testing.T) {
	input := `Device AA:BB:CC:DD:EE:FF (public)
	Name: My Headphones
	Alias: Headphones
	Class: 0x240418
	Icon: audio-headset
	Paired: yes
	Bonded: yes
	Trusted: yes
	Blocked: no
	Connected: yes
`
	d := parseDeviceInfo([]byte(input), testMACHeadphones)
	if d.MAC != testMACHeadphones {
		t.Errorf("MAC = %q, want %q", d.MAC, testMACHeadphones)
	}
	if d.Name != "My Headphones" {
		t.Errorf("Name = %q, want %q", d.Name, "My Headphones")
	}
	if !d.Connected {
		t.Error("Connected = false, want true")
	}
	if !d.Paired {
		t.Error("Paired = false, want true")
	}
	if !d.Trusted {
		t.Error("Trusted = false, want true")
	}
}

func TestParseDeviceInfoDisconnected(t *testing.T) {
	input := `Device AA:BB:CC:DD:EE:FF
	Name: Idle Device
	Paired: yes
	Trusted: no
	Connected: no
`
	d := parseDeviceInfo([]byte(input), testMACHeadphones)
	if d.Connected {
		t.Error("Connected = true, want false")
	}
	if !d.Paired {
		t.Error("Paired = false, want true")
	}
	if d.Trusted {
		t.Error("Trusted = true, want false")
	}
}

func TestParsePoweredStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{"powered on", "Controller AA:BB:CC:DD:EE:FF\n\tPowered: yes\n", true, false},
		{"powered off", "Controller AA:BB:CC:DD:EE:FF\n\tPowered: no\n", false, false},
		{"missing", "Controller AA:BB:CC:DD:EE:FF\n", false, true},
		{"empty", "", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePoweredStatus([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestValidateMAC(t *testing.T) {
	valid := []string{
		testMACHeadphones,
		"00:11:22:33:44:55",
		"ff:ee:dd:cc:bb:aa",
		"A0:B1:C2:D3:E4:F5",
	}
	for _, m := range valid {
		if err := validateMAC(m); err != nil {
			t.Errorf("validateMAC(%q) = %v, want nil", m, err)
		}
	}
	invalid := []string{
		"",
		"AA:BB:CC:DD:EE",
		"AA-BB-CC-DD-EE-FF",
		"GG:HH:II:JJ:KK:LL",
		"AA:BB:CC:DD:EE:FF; rm -rf /",
		"AABBCCDDEEFF",
		"AA:BB:CC:DD:EE:FF:00",
	}
	for _, m := range invalid {
		if err := validateMAC(m); err == nil {
			t.Errorf("validateMAC(%q) = nil, want error", m)
		}
	}
}

// TestGetDevicesWithMock exercises the runBluetoothctl seam end-to-end.
func TestGetDevicesWithMock(t *testing.T) {
	original := runBluetoothctl
	t.Cleanup(func() { runBluetoothctl = original })

	infoResponses := map[string]string{
		testMACHeadphones: "\tName: Headphones\n\tConnected: yes\n\tPaired: yes\n\tTrusted: yes\n",
		testMACMouse:      "\tName: Mouse\n\tConnected: no\n\tPaired: yes\n\tTrusted: no\n",
	}
	runBluetoothctl = func(_ context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "devices" {
			return []byte("Device " + testMACHeadphones + " Headphones\nDevice " + testMACMouse + " Mouse\n"), nil
		}
		if len(args) == 2 && args[0] == "info" {
			if body, ok := infoResponses[args[1]]; ok {
				return []byte(body), nil
			}
		}
		return nil, errors.New("unexpected call: " + strings.Join(args, " "))
	}

	devs, err := getDevices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 2 {
		t.Fatalf("got %d devices, want 2", len(devs))
	}

	byMAC := map[string]BluetoothDevice{}
	for _, d := range devs {
		byMAC[d.MAC] = d
	}
	if h := byMAC[testMACHeadphones]; !h.Connected || !h.Paired || !h.Trusted {
		t.Errorf("headphones state wrong: %+v", h)
	}
	if mo := byMAC[testMACMouse]; mo.Connected || !mo.Paired || mo.Trusted {
		t.Errorf("mouse state wrong: %+v", mo)
	}
}

func TestConnectDeviceValidatesMAC(t *testing.T) {
	called := false
	original := runBluetoothctlCombined
	t.Cleanup(func() { runBluetoothctlCombined = original })
	runBluetoothctlCombined = func(_ context.Context, _ ...string) ([]byte, error) {
		called = true
		return nil, nil
	}

	if err := connectDevice(context.Background(), "not-a-mac"); err == nil {
		t.Error("expected error for invalid MAC, got nil")
	}
	if called {
		t.Error("runBluetoothctlCombined should not be called with invalid MAC")
	}
}

func TestIsBluetoothEnabledMock(t *testing.T) {
	original := runBluetoothctl
	t.Cleanup(func() { runBluetoothctl = original })
	runBluetoothctl = func(_ context.Context, args ...string) ([]byte, error) {
		if len(args) != 1 || args[0] != "show" {
			return nil, errors.New("unexpected args")
		}
		return []byte("Controller AA:BB:CC:DD:EE:FF\n\tPowered: yes\n\tDiscoverable: no\n"), nil
	}

	on, err := isBluetoothEnabled(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !on {
		t.Error("expected enabled=true")
	}
}
