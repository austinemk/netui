package wifi

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// 1. Intercept for structural password overlays
	if m.UIState == StatePasswordInput {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				m.UIState = StateNormal
				m.PasswordInput = ""
			case "backspace":
				if len(m.PasswordInput) > 0 {
					m.PasswordInput = m.PasswordInput[:len(m.PasswordInput)-1]
				}
			case "enter":
				cmd := ConnectToAccessPoint(m.Client, m.SelectedAP, m.PasswordInput)
				m.UIState = StateNormal
				m.PasswordInput = ""
				return m, cmd
			default:
				if len(keyMsg.String()) == 1 {
					m.PasswordInput += keyMsg.String()
				}
			}
		}
		return m, nil
	}

	// 2. Intercept for Saved Network Profile Action Popups
	if m.UIState == StateSavedActionsMenu {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "up", "k":
				if m.MenuCursor > 0 {
					m.MenuCursor--
				}
			case "down", "j":
				if m.MenuCursor < 1 {
					m.MenuCursor++
				}
			case "esc":
				m.UIState = StateNormal
			case "enter":
				prof := m.Saved[m.Cursor]
				var cmd tea.Cmd
				if m.MenuCursor == 0 {
					cmd = ToggleAutoConnectCmd(m.Client, prof.UUID, !prof.AutoConnect)
				} else {
					cmd = ForgetProfileCmd(m.Client, prof.UUID)
				}
				m.UIState = StateNormal
				return m, cmd
			}
		}
		return m, nil
	}

	// 3. Normal State Core Navigation Loop
	switch msg := msg.(type) {
	case InfoLoadedMsg:
		m.Adapter = msg.Adapter
		m.Saved = msg.Saved
		m.ActiveAPs = msg.APs
		m.Loading = false
		if m.Scanning {
			return m, PollWifiTicker()
		}

	case ScanFinishedMsg:
		m.ActiveAPs = msg
		if m.Scanning {
			return m, PollWifiTicker()
		}

	case TickMsg:
		if m.Scanning {
			return m, TriggerHardwareScanCmd(m.Client)
		}

	case AdapterToggledMsg, ActionSuccessMsg:
		return m, func() tea.Msg {
			a, _ := GetAdapterSettings(m.Client)
			s, _ := GetSavedProfiles(m.Client)
			aps, _ := GetActiveAccessPoints(m.Client)
			return InfoLoadedMsg(InfoLoadedData{Adapter: a, Saved: s, APs: aps})
		}

	case ErrMsg:
		m.Err = msg
		m.Scanning = false

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			m.Scanning = !m.Scanning
			m.Cursor = 0
			if m.Scanning {
				return m, TriggerHardwareScanCmd(m.Client)
			}

		// Mode Actions when Scanning is Disabled
		case "p":
			if !m.Scanning {
				return m, TogglePowerCmd(m.Client, !m.Adapter.Enabled)
			}

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			max := len(m.Saved) - 1
			if m.Scanning {
				max = len(m.ActiveAPs) - 1
			}
			if m.Cursor < max {
				m.Cursor++
			}

		case "enter":
			if m.Scanning {
				if len(m.ActiveAPs) == 0 {
					return m, nil
				}
				m.SelectedAP = m.ActiveAPs[m.Cursor]

				// If profile matches a saved known config or is open, connect directly.
				if m.SelectedAP.Security == "NONE" || IsProfileSaved(m.Saved, m.SelectedAP.SSID) {
					return m, ConnectToAccessPoint(m.Client, m.SelectedAP, "")
				}
				// Else prompt for validation strings
				m.UIState = StatePasswordInput
			} else {
				if len(m.Saved) > 0 {
					m.UIState = StateSavedActionsMenu
					m.MenuCursor = 0
				}
			}
		}
	}
	return m, nil
}
