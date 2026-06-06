// Package bus for managing dbus
package bus

import "github.com/godbus/dbus/v5"

var conn *dbus.Conn

// Init opens the system D-Bus connection. Call once at startup.
func Init() error {
	c, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	conn = c
	return nil
}

// Get returns the shared system D-Bus connection.
func Get() *dbus.Conn {
	return conn
}

// Close cleanly shuts down the shared connection. Call once on exit.
func Close() {
	if conn != nil {
		conn.Close()
	}
}
