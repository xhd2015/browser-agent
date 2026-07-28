# Scenario

**Feature**: single connected session is notified and receives prepare_reconnect

```
Create sess-a; dial hello
BroadcastPrepareReconnect -> notified=["sess-a"]
  WS on sess-a receives type=prepare_reconnect with delay_ms
```

## Preconditions

- BroadcastCase = single-connected.
- SessionIDs = [sess-a]; ConnectSessionIDs = [sess-a].

## Steps

1. Set case and session ids.

## Context

- Minimal happy path for broadcast.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BroadcastCase = BroadcastSingleConnected
	req.SessionIDs = []string{"sess-a"}
	req.ConnectSessionIDs = []string{"sess-a"}
	req.BroadcastReason = "daemon-upgrade"
	req.BroadcastDelayMS = 1000
	return nil
}
```
