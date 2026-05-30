package wifi

// RenderSignal returns a Nerd Font Wi-Fi icon string based on the
// connection strength percentage and whether the network is secured or open.
func RenderSignal(strength uint8, security string) string {
	isSecured := security != "open" && security != ""

	switch {
	// Excellent (75% - 100%)
	case strength >= 75:
		if isSecured {
			return "󰤪" // Connected/Secured Lock-state Max
		}
		return "󰤨" // Open Max

	// Good (50% - 74%)
	case strength >= 50:
		if isSecured {
			return "󰤧"
		}
		return "󰤥"

	// Fair (25% - 49%)
	case strength >= 25:
		if isSecured {
			return "󰤤"
		}
		return "󰤢"

	// Weak (1% - 24%)
	case strength >= 1:
		if isSecured {
			return "󰤡"
		}
		return "󰤟"

	// No Signal / Dead (0%)
	default:
		if isSecured {
			return "󰤭" // Secured variant for zero/disconnected
		}
		return "󰤯" // Open/Empty outline
	}
}
