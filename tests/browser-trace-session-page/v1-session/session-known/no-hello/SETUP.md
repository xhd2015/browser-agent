# Scenario

**Feature**: no extension hello yet — waiting status JSON (requirement #1)

```
# Extension Agent has not contacted the server
Control Server Session: phase=waiting_extension, hello not received
Test Client -> GET /v1/session?session=<id>
Control Server -> {
  phase: waiting_extension,
  extension.connected: false,
  extension.supports_browser_trace: false,
  hint: non-empty waiting/install guidance
}
```

## Preconditions

- Live session id is known.
- No `POST /v1/hello` and no status posts before the probe.

## Steps

1. Set `DoHello = false` and `DoStatusRecording = false`.
2. Probe immediately after health becomes OK.

## Context

- Ready countdown fields should still be present (deadline/elapsed/remaining).
- Hint should mention waiting for the extension and/or install/enable guidance.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = false
	req.DoStatusRecording = false
	return nil
}
```
