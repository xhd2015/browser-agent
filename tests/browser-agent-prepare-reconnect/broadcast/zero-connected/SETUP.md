# Scenario

**Feature**: zero connected sockets → empty notified list, not an error

```
Create sess-x, sess-y; no WS dials
BroadcastPrepareReconnect -> notified=[]
```

## Preconditions

- BroadcastCase = zero-connected.
- SessionIDs present; ConnectSessionIDs empty.

## Steps

1. Set case with sessions but no connections.

## Context

- Admin path may still call broadcast when nothing is connected; response stays 200 with [].

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BroadcastCase = BroadcastZeroConnected
	req.SessionIDs = []string{"sess-x", "sess-y"}
	req.ConnectSessionIDs = nil
	req.BroadcastReason = "daemon-upgrade"
	req.BroadcastDelayMS = 1000
	return nil
}
```
