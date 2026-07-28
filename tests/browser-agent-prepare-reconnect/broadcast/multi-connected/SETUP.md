# Scenario

**Feature**: multi-session broadcast notifies every connected socket

```
Create sess-a, sess-b; both hello
BroadcastPrepareReconnect
  -> notified includes both
  -> both WS receive prepare_reconnect
```

## Preconditions

- BroadcastCase = multi-connected.
- SessionIDs = [sess-a, sess-b]; both connected.

## Steps

1. Set case and two session ids.

## Context

- Admin upgrade path must fan-out to all live extensions.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BroadcastCase = BroadcastMultiConnected
	req.SessionIDs = []string{"sess-a", "sess-b"}
	req.ConnectSessionIDs = []string{"sess-a", "sess-b"}
	req.BroadcastReason = "daemon-upgrade"
	req.BroadcastDelayMS = 1500
	return nil
}
```
