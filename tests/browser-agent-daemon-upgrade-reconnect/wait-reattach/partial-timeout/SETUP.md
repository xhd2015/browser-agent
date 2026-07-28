# Scenario

**Feature**: when deadline hits with partial reattach, stillWaiting holds missing ids

```
only sess-a reconnects; sess-b never -> reattached=[a], stillWaiting=[b]
```

## Preconditions

- Two-id waitList; mock always connects only sess-a.
- Short WaitTimeout so test finishes quickly.

## Steps

1. Configure partial-only mock and short timeout.

## Context

- Soft outcome for EnsureDaemon still-waiting warning path.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitReattachCase = WaitReattachPartialTimeout
	req.WaitList = []string{"sess-a", "sess-b"}
	req.MockFetchPlan = []MockFetchStep{
		{Sessions: []ConnSnap{
			{SessionID: "sess-a", Connected: true},
			{SessionID: "sess-b", Connected: false},
		}},
	}
	req.WaitTimeout = 40 * time.Millisecond
	req.PollInterval = 5 * time.Millisecond
	return nil
}
```
