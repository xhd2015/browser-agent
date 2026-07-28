# Scenario

**Feature**: reattach succeeds after several disconnected polls

```
poll1 disc, poll2 disc, poll3 both conn -> reattached full; FetchCalls>=3
```

## Preconditions

- WaitList one or two ids; MockFetchPlan delayed connection on 3rd step.

## Steps

1. Three-step plan ending in connected.

## Context

- Extensions need a few reconnect retries after daemon respawn.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitReattachCase = WaitReattachAllAfterPolls
	req.WaitList = []string{"sess-a", "sess-b"}
	disc := []ConnSnap{
		{SessionID: "sess-a", Connected: false},
		{SessionID: "sess-b", Connected: false},
	}
	conn := []ConnSnap{
		{SessionID: "sess-a", Connected: true},
		{SessionID: "sess-b", Connected: true},
	}
	req.MockFetchPlan = []MockFetchStep{
		{Sessions: disc},
		{Sessions: disc},
		{Sessions: conn},
	}
	req.WaitTimeout = 500 * time.Millisecond
	req.PollInterval = 5 * time.Millisecond
	return nil
}
```
