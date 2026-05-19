package bluetooth

import (
	"strings"
)

// DeviceIcon represents a specific Unicode glyph mapped from Nerd Fonts.
type DeviceIcon string

const (
	// Audio & Sound
	IconAudioHeadphones DeviceIcon = "󰋋" // Headphone Glyph
	IconAudioHeadset    DeviceIcon = "󰋎" // Headset with mic
	IconAudioCard       DeviceIcon = "󰓃" // Speaker / Audio source

	// Input Devices
	IconInputKeyboard DeviceIcon = "󰌌" // Keyboard
	IconInputMouse    DeviceIcon = "󰍽" // Mouse
	IconInputGaming   DeviceIcon = "󰊴" // Gamepad Controller
	IconInputTablet   DeviceIcon = "󰈬" // Drawing / Touch Tablet

	// Personal Electronics
	IconPhone       DeviceIcon = "󰏲" // Smartphone
	IconComputer    DeviceIcon = "󰟀" // Desktop Monitor / Computer
	IconCameraVideo DeviceIcon = "󰕧" // Video Camera
	IconCameraPhoto DeviceIcon = "󰄀" // Photo Camera
	IconSmartwatch  DeviceIcon = "󱎫" // Watch

	// Office & Smart Home
	IconPrinter         DeviceIcon = "󰐪" // Printer
	IconScanner         DeviceIcon = "󰚰" // Scanner
	IconModem           DeviceIcon = "󰖩" // Network Router / Modem
	IconNetworkWireless DeviceIcon = "󰖩" // Wireless AP
	IconDisplay         DeviceIcon = "󰍹" // Monitor display

	// Fallback
	IconGenericBluetooth DeviceIcon = "" // Bluetooth Symbol Logo
)

// String returns the string representation of the DeviceIcon.
func (i DeviceIcon) String() string {
	return string(i)
}

// FromString normalizes a raw icon string returned by BlueZ/bluetoothctl.
// It handles case-insensitivity and falls back gracefully to a generic icon glyph.
func FromString(iconStr string) DeviceIcon {
	switch strings.ToLower(iconStr) {
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

// FromClassOfDevice analyzes the CoD parameters provided by
// SIG Major Device Class specifications and assigns the closest matching Nerd Font glyph.
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
			switch minorPeripheral {
			case 0x01, 0x02, 0x03:
				return IconInputGaming // Joysticks, Gamepads
			case 0x05:
				return IconInputTablet // Digitizer Tablet
			default:
				return IconGenericBluetooth
			}
		}
	case 0x06: // Imaging (Cameras / Scanners)
		return IconCameraPhoto
	case 0x07: // Wearable (Smartwatches)
		return IconSmartwatch
	default:
		return IconGenericBluetooth
	}
}
