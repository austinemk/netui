package bluetooth

import (
	"github.com/godbus/dbus/v5"
)

const (
	agentInterface = "org.bluez.Agent1"
	agentPath      = "/corntui/bluetooth/agent"
)

// PasskeyAction encapsulates the user's decision context
type PasskeyAction struct {
	ResponseChan chan bool
}

type BluetoothAgent struct {
	// Channel used to send the pairing challenge to the Bubble Tea loop
	MsgChan chan<- PasskeyRequestMsg
}

// RequestConfirmation is called by BlueZ for Numeric Comparison pairing (e.g., "Is passkey 123456 correct?")
func (a *BluetoothAgent) RequestConfirmation(device dbus.ObjectPath, passkey uint32) *dbus.Error {
	logToFile("📲 Agent received RequestConfirmation for device %s, passkey: %06d", device, passkey)

	// Create a channel to wait for the Bubble Tea UI response
	responseChan := make(chan bool)

	// Construct a minimal Device object to show in the UI popup
	dev := Device{
		MAC:  string(device), // You can parse this or match it from FetchAllBlueZObjects
		Name: "Incoming Pairing Request",
	}

	// Ship the challenge over to Bubble Tea
	a.MsgChan <- PasskeyRequestMsg{
		Device:       dev,
		Passkey:      passkey,
		ResponseChan: responseChan,
	}

	// BLOCK here until the user clicks Yes or No in your Bubble Tea View
	accepted := <-responseChan

	if !accepted {
		logToFile("❌ User rejected the passkey confirmation")
		return dbus.NewError("org.bluez.Error.Rejected", []interface{}{"Pairing rejected by user"})
	}

	logToFile("✅ User accepted the passkey confirmation")
	return nil
}

// --- Required Agent1 Interface Stubs ---

func (a *BluetoothAgent) Release() *dbus.Error { return nil }

func (a *BluetoothAgent) RequestPinCode(device dbus.ObjectPath) (string, *dbus.Error) {
	return "", dbus.NewError("org.bluez.Error.Rejected", nil)
}

func (a *BluetoothAgent) DisplayPinCode(device dbus.ObjectPath, pincode string) *dbus.Error {
	return nil
}

func (a *BluetoothAgent) RequestPasskey(device dbus.ObjectPath) (uint32, *dbus.Error) {
	return 0, dbus.NewError("org.bluez.Error.Rejected", nil)
}

func (a *BluetoothAgent) DisplayPasskey(device dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	return nil
}

func (a *BluetoothAgent) AuthorizeService(device dbus.ObjectPath, uuid string) *dbus.Error {
	return nil
}
func (a *BluetoothAgent) Cancel() *dbus.Error { return nil }
