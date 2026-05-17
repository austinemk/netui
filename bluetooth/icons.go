package bluetooth

import (
	"strings"
)

// DeviceIcon represents the standardized freedesktop/BlueZ icon string type.
type DeviceIcon string

const (
	// Audio & Sound
	IconAudioHeadphones DeviceIcon = "audio-headphones"
	IconAudioHeadset    DeviceIcon = "audio-headset"
	IconAudioCard       DeviceIcon = "audio-card"

	// Input Devices
	IconInputKeyboard DeviceIcon = "input-keyboard"
	IconInputMouse    DeviceIcon = "input-mouse"
	IconInputGaming   DeviceIcon = "input-gaming"
	IconInputTablet   DeviceIcon = "input-tablet"

	// Personal Electronics
	IconPhone       DeviceIcon = "phone"
	IconComputer    DeviceIcon = "computer"
	IconCameraVideo DeviceIcon = "camera-video"
	IconCameraPhoto DeviceIcon = "camera-photo"
	IconSmartwatch  DeviceIcon = "smartwatch"

	// Office & Smart Home
	IconPrinter         DeviceIcon = "printer"
	IconScanner         DeviceIcon = "scanner"
	IconModem           DeviceIcon = "modem"
	IconNetworkWireless DeviceIcon = "network-wireless"
	IconDisplay         DeviceIcon = "display"

	// Fallback
	IconGenericBluetooth DeviceIcon = "bluetooth"
)

// String returns the string representation of the DeviceIcon.
func (i DeviceIcon) String() string {
	return string(i)
}

// FromString normalizes a raw icon string returned by BlueZ/bluetoothctl.
// It handles case-insensitivity and falls back gracefully to a generic icon.
func FromString(iconStr string) DeviceIcon {
	switch strings.ToLower(strings.TrimSpace(iconStr)) {
	case "audio-headphones":
		return IconAudioHeadphones
	case "audio-headset":
		return IconAudioHeadset
	case "audio-card":
		return IconAudioCard
	case "input-keyboard":
		return IconInputKeyboard
	case "input-mouse":
		return IconInputMouse
	case "input-gaming":
		return IconInputGaming
	case "input-tablet":
		return IconInputTablet
	case "phone":
		return IconPhone
	case "computer":
		return IconComputer
	case "camera-video":
		return IconCameraVideo
	case "camera-photo":
		return IconCameraPhoto
	case "smartwatch":
		return IconSmartwatch
	case "printer":
		return IconPrinter
	case "scanner":
		return IconScanner
	case "modem":
		return IconModem
	case "network-wireless":
		return IconNetworkWireless
	case "display":
		return IconDisplay
	default:
		return IconGenericBluetooth
	}
}

// FromClassOfDevice parses a raw Bluetooth CoD (Class of Device) uint32 bitmask.
// This matches the official Bluetooth SIG Major Device Class specifications.
func FromClassOfDevice(cod uint32) DeviceIcon {
	// Major Device Class is stored in bits 8 to 12 of the CoD
	majorClass := (cod >> 8) & 0x1F

	switch majorClass {
	case 0x01: // Computer
		return IconComputer
	case 0x02: // Phone
		return IconPhone
	case 0x03: // Network Access Point
		return IconNetworkWireless
	case 0x04: // Audio/Video
		// We can look at the minor class (bits 2 to 7) for specific sub-types
		minorClass := (cod >> 2) & 0x3F
		switch minorClass {
		case 0x01, 0x02: // Wearable Headset / Hands-free
			return IconAudioHeadset
		case 0x06: // Headphones
			return IconAudioHeadphones
		default:
			return IconAudioCard // Speakers, stereos, etc.
		}
	case 0x05: // Peripheral (Mouse, Keyboard, Joysticks)
		// Check bits 6 and 7 of the CoD to differentiate input types
		switch (cod >> 6) & 0x03 {
		case 0x01:
			return IconInputKeyboard
		case 0x02:
			return IconInputMouse
		case 0x03:
			return IconInputKeyboard // Combo keyboard/mouse defaults to keyboard
		default:
			// If not mouse/keyboard, check minor class bits for joysticks/gamepads
			minorPeripheral := (cod >> 2) & 0x0F
			if minorPeripheral == 0x01 || minorPeripheral == 0x02 {
				return IconInputGaming
			}
			return IconInputTablet
		}
	case 0x06: // Imaging (Cameras, Scanners)
		return IconCameraPhoto
	case 0x07: // Wearable (Smartwatches)
		return IconSmartwatch
	case 0x08: // Toy / Gaming
		return IconInputGaming
	default:
		return IconGenericBluetooth
	}
}
