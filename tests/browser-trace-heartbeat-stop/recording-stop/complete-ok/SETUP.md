# Scenario

**Feature**: normal complete after continuous status/entries (regression)

```
# Mock keeps heartbeat alive then completes successfully
Mock Extension -> recording status
Mock Extension -> POST status + entries repeatedly
Mock Extension -> POST /v1/complete
browser-trace -> exit 0, HAR + meta, no heartbeat_lost warning
```

## Preconditions

- HeartbeatTimeout may stay at product default (0 → 10s) or a generous test
  value so continuous ticks never trip heartbeat_lost.
- Complete arrives well before HeartbeatTimeout.

## Steps

1. Ensure HeartbeatTimeout is not short enough to race continuous ticks
   (leave 0 for product default, or set ≥ 5s).
2. Descendants set ExtContinuousComplete.

## Context

- Requirement scenarios #2 and #5.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Generous heartbeat so continuous-complete cannot trip heartbeat_lost.
	req.HeartbeatTimeout = 5 * time.Second
	req.ReadyTimeout = 3 * time.Second
	req.CompleteTimeout = 3 * time.Second
	return nil
}
```
