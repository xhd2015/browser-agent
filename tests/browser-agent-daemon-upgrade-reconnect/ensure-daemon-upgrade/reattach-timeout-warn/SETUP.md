# Scenario

**Feature**: when waitList never reconnects within ReattachWait, stderr warns still waiting; EnsureDaemon still succeeds

```
ReattachSessionIDs empty + short ReattachWait
  -> stderr warning still waiting + ids
  -> EnsureDaemon err nil (soft warn)
```

## Preconditions

- ConnectedSessionIDs non-empty; ReattachSessionIDs empty (never reconnect).
- ReattachWait very short (e.g. 30ms).

## Steps

1. Configure timeout reattach path.

## Context

- Product default wait is 30s; tests use short ReattachWait via config field.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureUpgradeCase = EnsureUpgradeReattachTimeoutWarn
	req.ConnectedSessionIDs = []string{"sess-w1", "sess-w2"}
	req.ReattachSessionIDs = nil // never reattach
	req.ReattachWait = 30 * time.Millisecond
	req.ClientVersion = "0.2.0"
	req.DaemonVersion = "0.1.0"
	return nil
}
```
