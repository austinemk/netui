# linktui

> A fast, minimal terminal UI for managing Wi-Fi, Bluetooth, and WireGuard VPN on Linux — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

---

## Screenshots

<!-- Replace with actual screenshots or a demo GIF -->

```
📸 Place your screenshots or GIF here
   Recommended: a single animated GIF (800×400px) showing tab switching
   Tools: ttyrec + gif-tty, asciinema, or vhs (https://github.com/charmbracelet/vhs)
```

| Wi-Fi                      | Bluetooth                            | VPN                      |
| -------------------------- | ------------------------------------ | ------------------------ |
| ![wifi](./assets/wifi.png) | ![bluetooth](./assets/bluetooth.png) | ![vpn](./assets/vpn.png) |

---

## Features

- **Wi-Fi** — scan nearby access points, connect with password prompt, manage saved profiles (autoconnect toggle, forget)
- **Bluetooth** — discover devices, pair with passkey confirmation, manage known devices, toggle power/discoverable/pairable
- **WireGuard VPN** — list tunnel profiles, activate/deactivate links, create new profiles via form, import `.conf` files via file picker, display public IP info
- Tab navigation with `Tab` / `Shift+Tab`
- Lazy-loads each tab on first visit for fast startup
- Fully themeable via `config.toml`
- Respects configured window dimensions — warns if terminal is too small

---

## Dependencies

### Runtime

| Dependency                                    | Purpose                         |
| --------------------------------------------- | ------------------------------- |
| [NetworkManager](https://networkmanager.dev/) | Wi-Fi and WireGuard VPN backend |
| [BlueZ](http://www.bluez.org/) (`bluetoothd`) | Bluetooth backend via D-Bus     |
| D-Bus                                         | IPC for BlueZ communication     |

Ensure NetworkManager and BlueZ daemons are running:

```sh
sudo systemctl enable --now NetworkManager
sudo systemctl enable --now bluetooth
```

### Build (Go modules)

| Module                                | Used for                                  |
| ------------------------------------- | ----------------------------------------- |
| `charm.land/bubbletea/v2`             | TUI framework                             |
| `charm.land/lipgloss/v2`              | Terminal styling                          |
| `charm.land/bubbles/v2`               | Table, text input, file picker components |
| `github.com/Wifx/gonetworkmanager/v3` | NetworkManager Go bindings                |
| `github.com/godbus/dbus/v5`           | D-Bus bindings for BlueZ                  |
| `github.com/BurntSushi/toml`          | Config file parsing                       |

---

## Installation

### AUR (Arch Linux)

```sh
# Using yay
yay -S linktui

# Using paru
paru -S linktui

# Manually
git clone https://aur.archlinux.org/linktui.git
cd linktui
makepkg -si
```

### GitHub Releases (pre-built binary)

Download the latest binary from the [Releases page](https://github.com/yourusername/linktui/releases).

```sh
# Download and install (replace VERSION and ARCH as needed)
curl -Lo linktui https://github.com/yourusername/linktui/releases/latest/download/linktui-linux-amd64
chmod +x linktui
sudo mv linktui /usr/local/bin/
```

### Build from source

Requires Go 1.22+.

```sh
git clone https://github.com/yourusername/linktui.git
cd linktui
go build -o linktui ./cmd/linktui
sudo mv linktui /usr/local/bin/
```

---

## Usage

```sh
# Open on the Wi-Fi tab (default)
linktui

# Open directly on a specific tab
linktui --tab wifi
linktui --tab bluetooth
linktui --tab vpn
```

### Keybindings

| Key                 | Action                |
| ------------------- | --------------------- |
| `Tab` / `Shift+Tab` | Switch tabs           |
| `j` / `k`           | Navigate list up/down |
| `q`                 | Quit                  |
| `Ctrl+C`            | Force quit            |

**Wi-Fi tab**

| Key     | Action                                     |
| ------- | ------------------------------------------ |
| `s`     | Toggle scan mode                           |
| `Enter` | Connect (scan mode) / Options (saved mode) |
| `p`     | Toggle adapter power                       |

**Bluetooth tab**

| Key     | Action                    |
| ------- | ------------------------- |
| `s`     | Start/stop discovery scan |
| `Enter` | Device actions menu       |
| `p`     | Toggle power              |
| `d`     | Toggle discoverable       |
| `b`     | Toggle pairable           |

**VPN tab**

| Key     | Action                                    |
| ------- | ----------------------------------------- |
| `n`     | Create new WireGuard profile              |
| `i`     | Import `.conf` file                       |
| `Enter` | Actions menu (activate/deactivate/delete) |
| `p`     | Fetch public IP info                      |

---

## Configuration

linktui looks for a config file at:

```
~/.config/linktui/config.toml
```

If the file is absent, built-in defaults are used. All values are optional — only override what you need.

### Example `config.toml`

```toml
[window]
width  = 80
height = 28

[colors]
foreground           = "#CDD6F4"
background           = "#1E1E2E"
accent               = "#89B4FA"
highlight            = "#CBA6F7"
highlight_background = "#313244"
muted                = "#6C7086"
border               = "#45475A"
popup_background     = "#181825"
log_background       = "#11111B"
cursor               = "#F5E0DC"
```

> **Note:** `width` and `height` set the UI canvas size, not the terminal window size. The terminal must be at least as large as the configured dimensions — linktui will display a warning if it isn't.

### Minimum dimensions

| Setting  | Minimum |
| -------- | ------- |
| `width`  | 70      |
| `height` | 25      |

---

## Permissions

linktui communicates with system daemons over D-Bus and does not require `sudo`. Your user must be in the appropriate groups or have polkit policies that allow NetworkManager and BlueZ interactions.

On most distributions, logging in via a desktop session handles this automatically. If you run into permission errors in a minimal/headless setup:

```sh
# Add your user to the 'network' group if required by your distro
sudo usermod -aG network $USER
```

---

## License

MIT — see [LICENSE](./LICENSE) for details.
