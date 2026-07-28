# Scenario

**Feature**: first poll shows all waitList ids connected → full reattach

```
fetch #1: all connected -> reattached=waitList, stillWaiting=[]
```

## Preconditions

- WaitList has two ids; first MockFetchStep connects both.

## Steps

1. Configure immediate-success mock fetch.

## Context

- Happy path after new daemon accepts extension hellos quickly.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitReattachCase = WaitReattachAllImmediate
	req.WaitList = []string{"sess-a", "sess-b"}
	req.MockFetchPlan = []MockFetchStep{
		{Sessions: []ConnSnap{
			{SessionID: "sess-a", Connected: true},
			{SessionID: "sess-b", Connected: true},
		}},
	}
	req.WaitTimeout = 200 * time.Millisecond
	req.PollInterval = 5 * time.Millisecond
	return nil
}
```
