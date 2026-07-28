# Scenario

**Feature**: transient fetch errors are retried until a successful connected poll

```
fetch err, fetch err, then all connected -> reattached full (no hard fail)
```

## Preconditions

- WaitList one id; first two steps error; third connects.

## Steps

1. MockFetchPlan with FetchErr then success.

## Context

- New daemon may briefly refuse GET /v1/sessions during boot.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitReattachCase = WaitReattachFetchErrorRetries
	req.WaitList = []string{"sess-a"}
	req.MockFetchPlan = []MockFetchStep{
		{FetchErr: "connection refused"},
		{FetchErr: "temporary unavailable"},
		{Sessions: []ConnSnap{
			{SessionID: "sess-a", Connected: true},
		}},
	}
	req.WaitTimeout = 500 * time.Millisecond
	req.PollInterval = 5 * time.Millisecond
	return nil
}
```
