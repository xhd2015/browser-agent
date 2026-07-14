# Scenario

**Feature**: full recording session stop paths (heartbeat_lost vs complete)

```
# CLI runs sealed session with mock extension (no Chrome)
User -> browser-trace (NoOpenChrome, HeartbeatTimeout injectable)
Mock Extension -> POST /v1/hello
Mock Extension -> GET /v1/commands (start)
Mock Extension -> POST /v1/status (recording)
Mock Extension ?-> POST /v1/entries (snapshot)
# Either silence → heartbeat_lost, or continuous + complete
```

## Preconditions

- Mode is full session lifecycle (`ModeSession`).
- Ready must succeed (hello + recording status) before stop paths apply.
- Injectable `HeartbeatTimeout` is used by heartbeat_lost descendants.

## Steps

1. Set `Mode = ModeSession`.
2. Default `SessionSuffix` for mock entries session_id alignment.
3. Descendants set ExtensionScript and HeartbeatTimeout / complete behavior.

## Context

- Requirement scenarios #2–#5.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSession
	if req.SessionSuffix == "" {
		req.SessionSuffix = "hb-session"
	}
	if req.MockWindowID == 0 {
		req.MockWindowID = 42
	}
	return nil
}
```
