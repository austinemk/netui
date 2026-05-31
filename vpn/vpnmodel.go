package vpn

import (
	"os"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/table"
)

type Model struct {
	Client     *DBusClient
	Tunnels    []TunnelProfile
	Table      table.Model
	FilePicker filepicker.Model // Integrated Native File Picker Component
	MenuCursor int
	UIState    UIState
	Loading    bool
	Err        error
	Height     int // Terminal height for filepicker sizing

	// Form input states
	ActiveField FormField
	FormInputs  map[FormField]string
}

func New() Model {
	// 1. Initialize table view layout components
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 15},
		{Title: "Status", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// 2. Instantiate the file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".conf", ".wg", ".jpg", ".png"}

	// Start at home directory; fall back to cwd if home is unavailabe
	fp.CurrentDirectory, _ = os.UserHomeDir()

	return Model{
		Client:     &DBusClient{NM: nil},
		Table:      t,
		FilePicker: fp,
		Loading:    true,
		UIState:    StateNormal,
		FormInputs: make(map[FormField]string),
	}
}
