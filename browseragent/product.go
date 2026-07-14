// Package browseragent implements the browser-agent control plane:
// session resolve, job queue/RPC, WebSocket extension control, and session UI.
package browseragent

const (
	// DefaultAddr is the product control listen address.
	DefaultAddr = "127.0.0.1:43761"
	// DefaultControlPort is the product control port (string for HTML/CLI markers).
	DefaultControlPort = "43761"
	// ProductName is the CLI / product display name.
	ProductName = "browser-agent"
	// FeatureBrowserAgent is the capability advertised by the extension hello.
	FeatureBrowserAgent = "browser-agent"
	// MinBrowserAgentVersion is the floor for supports_browser_agent.
	MinBrowserAgentVersion = "1.0.0"

	// Session phases
	PhaseWaitingExtension    = "waiting_extension"
	PhaseExtensionConnected  = "extension_connected"

	// Job statuses
	JobStatusQueued  = "queued"
	JobStatusRunning = "running"
	JobStatusDone    = "done"
	JobStatusFailed  = "failed"
	JobStatusExpired = "expired"

	// Canonical job type strings (shared with CLI / extension / react protocol).
	JobTypeInfo       = "info"
	JobTypeEval       = "eval"
	JobTypeRun        = "run"
	JobTypeLogs       = "logs"
	JobTypeScreenshot = "screenshot"
	JobTypeCDP        = "cdp"
)

// IsKnownJobType reports whether s is one of the six canonical job type strings
// (exact lowercase match).
func IsKnownJobType(s string) bool {
	switch s {
	case JobTypeInfo, JobTypeEval, JobTypeRun, JobTypeLogs, JobTypeScreenshot, JobTypeCDP:
		return true
	default:
		return false
	}
}
