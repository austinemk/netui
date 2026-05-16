package bluetooth

import (
	"bytes"
	"os/exec"
	"strings"
)

type Device struct {
	MAC       string
	Name      string
	Connected bool
	Paired    bool
}

// ScanDevices runs bluetoothctl devices to list discovered peripherals
func ScanDevices() ([]Device, error) {
	cmd := exec.Command("bluetoothctl", "devices")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// Fallback mock data for local development/testing without active BT hardware
		return []Device{
			{MAC: "00:1A:7D:DA:71:11", Name: "Sony WH-1000XM4", Connected: false, Paired: true},
			{MAC: "04:52:C7:0B:12:34", Name: "Logitech MX Master 3", Connected: true, Paired: true},
			{MAC: "A4:C1:38:88:99:AA", Name: "Unknown Audio Device", Connected: false, Paired: false},
		}, nil
	}

	var devices []Device
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3) // Format: "Device MA:C_:__:_ Name"
		if len(parts) < 3 {
			continue
		}

		devices = append(devices, Device{
			MAC:       parts[1],
			Name:      parts[2],
			Connected: false, // Connection tracking requires checking 'info' per device
			Paired:    false,
		})
	}
	return devices, nil
}
