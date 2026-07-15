// Package browseragent implements the browser-agent control plane:
// session resolve, job queue/RPC, WebSocket extension control, and session UI.
package browseragent

const (
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
	JobTypeCreateTab  = "create_tab"
)

// IsKnownJobType reports whether s is a canonical job type string
// (exact lowercase match). Additive set includes prior six plus create_tab.
func IsKnownJobType(s string) bool {
	switch s {
	case JobTypeInfo, JobTypeEval, JobTypeRun, JobTypeLogs, JobTypeScreenshot, JobTypeCDP, JobTypeCreateTab:
		return true
	default:
		return false
	}
}
