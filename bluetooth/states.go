package bluetooth

// AppState represents the overarching feature/view the user is looking at
type AppState int

const (
	StateBluetooth AppState = iota
	StateWiFi
	StateVPN
)

// OpStatus represents what a specific sub-system is actively doing
type OpStatus int

const (
	StatusIdle       OpStatus = iota
	StatusScanning            // Used for BT, Wi-Fi, or VPN discovery
	StatusConnecting          // Active handshake/connection phase
	StatusConnected           // Successfully established connection
	StatusDisconnecting
	StatusError
)

// String helper methods make it incredibly easy to render status text in Lipgloss
func (s OpStatus) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusScanning:
		return "Scanning..."
	case StatusConnecting:
		return "Connecting..."
	case StatusConnected:
		return "Connected"
	case StatusDisconnecting:
		return "Disconnecting..."
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}
