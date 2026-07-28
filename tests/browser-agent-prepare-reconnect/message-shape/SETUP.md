# Scenario

**Feature**: pure prepare_reconnect payload builder (message shape)

```
Test Client -> BuildPrepareReconnectPayload(opts)
  -> map with delay_ms (default 1000), optional reason / retry hints
```

## Preconditions

- Mode is `message-shape`.
- No HTTP server; no WebSocket.
- Package API `browseragent.BuildPrepareReconnectPayload` must exist (RED until implementer).

## Steps

1. Set `Mode = ModeMessageShape`.
2. Children set `MessageShapeCase` and payload option fields.

## Context

- WS envelope type string is `prepare_reconnect` (asserted at broadcast/HTTP layers).
- This branch only checks **payload** shape from the pure builder.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeMessageShape
	return nil
}
```
