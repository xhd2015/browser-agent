# Scenario

**Feature**: heartbeat_lost after POST /v1/entries snapshot with N URLs

```
# Mock records, pushes snapshot, then goes silent
Mock Extension -> POST /v1/hello
Mock Extension -> wait start command
Mock Extension -> POST /v1/status recording
Mock Extension -> POST /v1/entries {urls: alpha, app.js}
Mock Extension -> silence > HeartbeatTimeout

# Server saves partial session
Control Server -> recording.har contains snapshot URLs
Control Server -> meta stop_reason=heartbeat_lost partial=true
browser-trace -> exit 0, stderr warning, stdout path\n
```

## Preconditions

- Snapshot URLs are deterministic fixture strings for HAR substring asserts.
- HeartbeatTimeout already short from parent grouping.

## Steps

1. Set `ExtensionScript = ExtSilenceWithSnapshot`.
2. Set `SnapshotURLs` to two known example URLs.

## Context

- Requirement scenario #3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtSilenceWithSnapshot
	req.SnapshotURLs = []string{
		"https://api.example.com/v1/alpha",
		"https://cdn.example.com/app.js",
	}
	return nil
}
```
