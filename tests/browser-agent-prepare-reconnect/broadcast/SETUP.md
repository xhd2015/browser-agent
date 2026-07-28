# Scenario

**Feature**: BroadcastPrepareReconnect notifies connected extension WebSockets only

```
# create sessions; dial subset with hello
Fake Extension -> /v1/ws?session=S -> hello
Test Client -> BroadcastPrepareReconnect(registry, opts)
  -> type=prepare_reconnect on each connected socket
  -> returns notified session ids (write-ok only)
```

## Preconditions

- Mode is `broadcast`.
- Uses `NewSessionRegistry` + `NewRegistryControlHandler` + httptest.
- Package API `browseragent.BroadcastPrepareReconnect` must exist (RED until implementer).

## Steps

1. Set `Mode = ModeBroadcast`.
2. Children set BroadcastCase, SessionIDs, ConnectSessionIDs, reason/delay.

## Context

- Disconnected sessions must not appear in notified.
- Empty notified when no sockets is success (not an error).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeBroadcast
	if req.BroadcastReason == "" {
		req.BroadcastReason = "daemon-upgrade"
	}
	if req.BroadcastDelayMS == 0 {
		req.BroadcastDelayMS = 1000
	}
	return nil
}
```
