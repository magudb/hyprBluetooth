# HyprBluetooth

A terminal-based Bluetooth device manager built with Go and Bubble Tea, designed for Linux systems and Hyprland users.

## Features

- 🔵 **Interactive TUI**: Beautiful terminal interface with mouse and keyboard support
- 📱 **Device Management**: Scan, pair, connect, and disconnect Bluetooth devices
- ⚡ **Real-time Status**: Live updates of device connection states
- 🎛️ **Bluetooth Control**: Enable/disable Bluetooth adapter
- 🖱️ **Mouse Support**: Full mouse interaction including scrolling and clicking
- ⌨️ **Keyboard Navigation**: Vim-style navigation keys

## Screenshots

```
┌─────────────────────────────────────────────────┐
│ HyprBluetooth - Bluetooth Device Manager        │
│ 🔵 Bluetooth: ON                                │
│                                                 │
│ > ● WH-1000XM4 (00:11:22:33:44:55)             │
│   ◐ Magic Mouse (66:77:88:99:AA:BB)            │
│   ○ Unknown Device (CC:DD:EE:FF:00:11)         │
│                                                 │
│ Controls:                                       │
│   ↑/k, ↓/j: Navigate  Enter/Space: Connect     │
│   s: Scan  r: Refresh  p: Pair  d: Disconnect  │
│   e: Enable/Disable Bluetooth  q: Quit         │
│                                                 │
│ Status: ● Connected  ◐ Paired  ○ Unpaired      │
└─────────────────────────────────────────────────┘
```

## Installation

### From Release (Recommended)

Download the latest release for your architecture:

```bash
# For AMD64
curl -LO https://github.com/yourusername/hyprBluetooth/releases/latest/download/hyprBluetooth_linux_amd64.tar.gz
tar -xzf hyprBluetooth_linux_amd64.tar.gz
sudo mv hyprBluetooth /usr/local/bin/

# For ARM64
curl -LO https://github.com/yourusername/hyprBluetooth/releases/latest/download/hyprBluetooth_linux_arm64.tar.gz
tar -xzf hyprBluetooth_linux_arm64.tar.gz
sudo mv hyprBluetooth /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/yourusername/hyprBluetooth.git
cd hyprBluetooth
go build -o hyprBluetooth .
sudo mv hyprBluetooth /usr/local/bin/
```

## Prerequisites

- Linux system with BlueZ stack
- `bluetoothctl` command available
- Go 1.24+ (for building from source)

### Installing Dependencies

**Arch Linux:**
```bash
sudo pacman -S bluez bluez-utils
sudo systemctl enable --now bluetooth
```

**Ubuntu/Debian:**
```bash
sudo apt install bluetooth bluez bluez-tools
sudo systemctl enable --now bluetooth
```

**Fedora:**
```bash
sudo dnf install bluez bluez-tools
sudo systemctl enable --now bluetooth
```

## Usage

Run the application:

```bash
hyprBluetooth
```

### Controls

| Key | Action |
|-----|--------|
| `↑/k` | Move cursor up |
| `↓/j` | Move cursor down |
| `Enter/Space` | Connect/disconnect selected device |
| `s` | Scan for new devices |
| `r` | Refresh device list |
| `p` | Pair selected device |
| `d` | Disconnect selected device |
| `e` | Enable/disable Bluetooth adapter |
| `Ctrl+r` | Full refresh (devices + Bluetooth status) |
| `q/Ctrl+c` | Quit application |

### Mouse Support

- **Scroll wheel**: Navigate up/down through device list
- **Left click**: Select device
- **All controls**: Fully functional with mouse

### Device Status Indicators

- `●` **Connected**: Device is actively connected
- `◐` **Paired**: Device is paired but not connected
- `○` **Unpaired**: Device is discovered but not paired

## Configuration

hyprBluetooth works out of the box with no configuration required. It uses the system's BlueZ stack through `bluetoothctl` commands.

## Integration with Hyprland

You can bind hyprBluetooth to a key combination in your Hyprland config:

```conf
# ~/.config/hypr/hyprland.conf
bind = SUPER, B, exec, hyprBluetooth
```

Or create a floating window rule:

```conf
windowrule = float, ^(hyprBluetooth)$
windowrule = size 800 600, ^(hyprBluetooth)$
windowrule = center, ^(hyprBluetooth)$
```

## Troubleshooting

### Bluetooth service not running
```bash
sudo systemctl start bluetooth
sudo systemctl enable bluetooth
```

### Permission issues
Make sure your user is in the bluetooth group:
```bash
sudo usermod -a -G bluetooth $USER
# Log out and log back in
```

### Command not found
Ensure `bluetoothctl` is installed and in your PATH:
```bash
which bluetoothctl
```

### Device won't connect
1. Try unpairing and re-pairing the device
2. Make sure the device is in pairing mode
3. Check if the device is already connected to another system

## Development

### Building

```bash
go mod download
go build -o hyprBluetooth .
```

### Running Tests

```bash
go test ./...
```

### Linting

```bash
golangci-lint run
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Inspired by the need for a simple Bluetooth manager for tiling window managers

## Related Projects

- [bluetuith](https://github.com/darkhz/bluetuith) - Another TUI Bluetooth manager
- [blueman](https://github.com/blueman-project/blueman) - GTK+ Bluetooth manager
- [blueberry](https://github.com/linuxmint/blueberry) - Linux Mint's Bluetooth configuration tool