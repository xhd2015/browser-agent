# Scenario

**Feature**: capable hello, not yet recording (requirement #2)

```
# Hello OK; no recording status yet
Test Client -> POST /v1/hello {version:1.2.0, features:[browser-trace,…]}
Test Client -> GET /v1/session
Control Server -> {
  phase: extension_connected,
  extension.connected: true,
  extension.supports_browser_trace: true,
  extension.version: "1.2.0",
  recording.active: false
}
```

## Preconditions

- Capability-satisfying hello already configured by parent Setup.
- No `POST /v1/status` with recording before the probe.

## Steps

1. Set `DoStatusRecording = false`.
2. Probe session JSON after hello settles.

## Context

- Phase should be the intermediate UI phase `extension_connected` (not
  `waiting_extension`, not `recording`).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoStatusRecording = false
	return nil
}
```
