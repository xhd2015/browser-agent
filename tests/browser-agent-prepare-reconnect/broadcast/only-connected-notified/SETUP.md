# Scenario

**Feature**: disconnected sessions are skipped; only connected id is notified

```
Create sess-connected + sess-orphan; only connected dials WS
BroadcastPrepareReconnect
  -> notified=["sess-connected"] only
  -> orphan never appears
```

## Preconditions

- BroadcastCase = only-connected-notified.
- SessionIDs = [sess-connected, sess-orphan].
- ConnectSessionIDs = [sess-connected] only.

## Steps

1. Set case and mixed connectivity.

## Context

- Matches upgrade policy: only live extensions need prepare_reconnect.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BroadcastCase = BroadcastOnlyConnectedNotified
	req.SessionIDs = []string{"sess-connected", "sess-orphan"}
	req.ConnectSessionIDs = []string{"sess-connected"}
	req.BroadcastReason = "daemon-upgrade"
	req.BroadcastDelayMS = 1000
	return nil
}
```
