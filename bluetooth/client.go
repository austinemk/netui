package bluetooth

/*import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	ErrDeviceMacEmpty  = errors.New("device MAC address cannot be empty")
	ErrExecutionFailed = errors.New("bluetoothctl command execution failed")
)

// executeBluetoothctlCmd is a private helper to run explicit subcommands
// natively through the bluetoothctl CLI wrapper.
func executeBluetoothctlCmd(subcmd string, mac string) (string, error) {
	if strings.TrimSpace(mac) == "" {
		return "", ErrDeviceMacEmpty
	}

	// This executes as: bluetoothctl pair 00:11:22:33:44:55
	cmd := exec.Command("bluetoothctl", subcmd, mac)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s (system err: %v)", ErrExecutionFailed, stderr.String(), err)
	}

	return stdout.String(), nil
}

// Pair initiates a secure cryptographic key-exchange pairing process with a remote device.
func Pair(mac string) error {
	output, err := executeBluetoothctlCmd("pair", mac)
	if err != nil {
		return err
	}

	// Even if the terminal command returns 0 exit, check for internal protocol failure strings
	if strings.Contains(output, "Failed to pair") {
		return fmt.Errorf("pairing protocol rejected: %s", strings.TrimSpace(output))
	}
	return nil
}

// Connect attempts to actively establish a live RF communication stream (Audio sink, HID profile, etc)
func Connect(mac string) error {
	output, err := executeBluetoothctlCmd("connect", mac)
	if err != nil {
		return err
	}

	if strings.Contains(output, "Failed to connect") {
		return fmt.Errorf("connection establishment failed: %s", strings.TrimSpace(output))
	}
	return nil
}

// Disconnect breaks an active live connection stream while maintaining paired status keys.
func Disconnect(mac string) error {
	output, err := executeBluetoothctlCmd("disconnect", mac)
	if err != nil {
		return err
	}

	if strings.Contains(output, "Failed to disconnect") {
		return fmt.Errorf("disconnection process rejected: %s", strings.TrimSpace(output))
	}
	return nil
}

// Trust whitelists a device so it can automatically reconnect to your computer in the future
func Trust(mac string) error {
	output, err := executeBluetoothctlCmd("trust", mac)
	if err != nil {
		return err
	}

	if strings.Contains(output, "Failed to trust") {
		return fmt.Errorf("failed to register trust status: %s", strings.TrimSpace(output))
	}
	return nil
}

// Distrust removes automatic reconnection privileges from a device
func Distrust(mac string) error {
	output, err := executeBluetoothctlCmd("untrust", mac)
	if err != nil {
		return err
	}

	if strings.Contains(output, "Failed to untrust") {
		return fmt.Errorf("failed to strip trust status: %s", strings.TrimSpace(output))
	}
	return nil
}

// Remove completely unpairs, forgets, and purges the device configuration from BlueZ storage
func Remove(mac string) error {
	output, err := executeBluetoothctlCmd("remove", mac)
	if err != nil {
		return err
	}

	if strings.Contains(output, "Failed to remove") {
		return fmt.Errorf("device deletion failed: %s", strings.TrimSpace(output))
	}
	return nil
}*/
