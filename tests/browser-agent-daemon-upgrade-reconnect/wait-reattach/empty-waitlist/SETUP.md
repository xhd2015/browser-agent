# Scenario

**Feature**: empty waitList returns immediately without requiring successful fetches

```
WaitSessionsReconnected([], fetch, ...) -> reattached=[], stillWaiting=[]
```

## Preconditions

- WaitReattachCase = empty-waitlist.
- WaitList empty; MockFetchPlan may be empty.

## Steps

1. Set empty WaitList and short timeout.

## Context

- Upgrade with zero connected sessions skips reattach work.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitReattachCase = WaitReattachEmptyWaitlist
	req.WaitList = nil
	req.MockFetchPlan = nil
	req.WaitTimeout = 50 * time.Millisecond
	req.PollInterval = 5 * time.Millisecond
	return nil
}
```
