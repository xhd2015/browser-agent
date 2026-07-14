package browseragent

// ComputeSessionStatus derives status and a human label from page count and extension state.
// pageCount nil means unknown telemetry (display —).
func ComputeSessionStatus(pageCount *int, connected, supportsBA bool) (status, label string) {
	if pageCount == nil {
		return "unknown", "Unknown"
	}
	count := *pageCount
	if count == 0 {
		return "no_session_page", "No session page"
	}
	if count > 1 {
		return "multiple_pages", "Multiple pages"
	}
	if !connected {
		return "page_no_extension", "Page open, no extension"
	}
	if supportsBA {
		return "ready", "Ready"
	}
	return "unsupported_extension", "Unsupported extension"
}

func statusColor(colors serveColor, status string, label string) string {
	switch status {
	case "no_session_page":
		return colors.red(label)
	case "multiple_pages", "unsupported_extension":
		return colors.orange(label)
	case "page_no_extension":
		return colors.yellow(label)
	case "ready":
		return colors.green(label)
	default:
		return colors.gray(label)
	}
}