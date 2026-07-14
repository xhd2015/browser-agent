# Scenario

**Feature**: continuous heartbeats then POST /v1/complete (normal success)

```
# Continuous liveness then extension complete
Mock Extension -> POST /v1/hello
Mock Extension -> wait start
Mock Extension -> POST /v1/status recording
Mock Extension -> loop: POST /v1/entries + POST /v1/status
Mock Extension -> POST /v1/complete (stop_reason=extension)

# Normal success contract
browser-trace -> exit 0
browser-trace stdout -> session path\n
browser-trace stderr -> no heartbeat_lost warning required
Storage -> recording.har + meta (not partial heartbeat_lost)
```

## Preconditions

- ContinuousTicks ≥ 2 so multiple heartbeats land before complete.
- HeartbeatTimeout from parent is large enough (5s).

## Steps

1. Set `ExtensionScript = ExtContinuousComplete`.
2. Set `ContinuousTicks = 3`.
3. Set `MockStopReason = extension`.
4. Optional SnapshotURLs for HAR content check.

## Context

- Requirement scenarios #2 (continuous + complete) and #5 (explicit complete still works).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtContinuousComplete
	req.ContinuousTicks = 3
	req.MockStopReason = "extension"
	req.SnapshotURLs = []string{"https://api.example.com/live"}
	return nil
}
```
