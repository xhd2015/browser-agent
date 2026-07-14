# Scenario

**Feature**: after status recording, session JSON reflects capture (requirement #4)

```
# Capable agent hellos then reports recording with entry_count
Test Client -> POST /v1/hello {version:1.2.0, features:[browser-trace,…]}
Test Client -> POST /v1/status {state: recording, entry_count: 7, window_id: 42}
Test Client -> GET /v1/session
Control Server -> {
  phase: recording,
  recording.active: true,
  recording.entry_count: 7,
  extension.supports_browser_trace: true
}
```

## Preconditions

- Parent sets capable hello (feature + version ≥ 1.2.0).
- Status recording is posted after hello before the probe.

## Steps

1. Set `DoStatusRecording = true`.
2. Set `EntryCount = 7` (non-zero, easy to assert).
3. Set `WindowID = 42` (default mock window).

## Context

- When recording is active, phase must be `recording` (not `extension_connected`).
- Supports flag remains true from hello.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoStatusRecording = true
	req.EntryCount = 7
	req.WindowID = 42
	return nil
}
```
