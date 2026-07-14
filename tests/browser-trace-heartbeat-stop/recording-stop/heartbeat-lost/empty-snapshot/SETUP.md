# Scenario

**Feature**: heartbeat_lost with no POST /v1/entries (empty/minimal HAR)

```
# Mock records then goes silent without ever pushing entries
Mock Extension -> POST /v1/hello
Mock Extension -> wait start command
Mock Extension -> POST /v1/status recording
Mock Extension -> silence (no /v1/entries) > HeartbeatTimeout

# Server still exits success-like with empty snapshot HAR
Control Server -> recording.har with empty log.entries
Control Server -> meta stop_reason=heartbeat_lost partial=true entry_count~0
browser-trace -> exit 0, stderr warning, stdout path\n
```

## Preconditions

- HeartbeatTimeout already short from parent grouping.
- previewEntries remains empty / nil until heartbeat_lost save.

## Steps

1. Set `ExtensionScript = ExtSilenceEmpty`.
2. Leave SnapshotURLs empty.

## Context

- Requirement scenario #4 — prefer exit 0 + warning + empty HAR + partial meta.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtSilenceEmpty
	req.SnapshotURLs = nil
	return nil
}
```
