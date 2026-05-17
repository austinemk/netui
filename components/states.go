package components

// OpStatus represents what any network sub-system is actively doing.
type OpStatus int

const (
	StatusIdle OpStatus = iota
	StatusScanning
	StatusConnecting
	StatusConnected
	StatusDisconnecting
	StatusError
)

// String helper method makes it easy for any view to output text
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
