package bluetooth

import (
	"github.com/godbus/dbus/v5"
)

type Device struct {
	MAC       string
	Name      string
	Connected bool
	Paired    bool
	Trusted   bool
	Icon      string
}

const (
	bluezInterface = "org.bluez"
	adapterPath    = "/org/bluez/hci0" // Default bluetooth adapter path
)

// Helper to grab a quick handle on the System Bus
func getSystemBus() (*dbus.Conn, error) {
	return dbus.ConnectSystemBus()
}
