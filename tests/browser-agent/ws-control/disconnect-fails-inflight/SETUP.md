# Scenario

**Feature**: WS disconnect fails inflight jobs (D4, policy v1)

```
Fake Extension hello; POST /v1/jobs starts (job may be running/queued)
Extension WS drops (no result)
  -> HTTP waiter unblocks with ok=false; error mentions disconnect/connection
DSN: fail inflight on disconnect — no requeue
```

## Preconditions

- WSAction disconnect-inflight.
- Extension does **not** auto-complete.

## Steps

1. Set WSAction to disconnect-fails-inflight.

## Context

- v1 simplicity: **fail** (not requeue-once). Documented in root DSN.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSAction = WSActionDisconnectInflight
	return nil
}
```
