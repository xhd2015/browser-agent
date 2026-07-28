# Scenario

**Feature**: WaitSessionsReconnected polls mock fetch until waitList reattached or timeout

```
WaitSessionsReconnected(waitList, mockFetch, timeout, pollInterval)
  -> reattached[], stillWaiting[]
```

## Preconditions

- Mode = `wait-reattach`.
- Leaf sets `WaitReattachCase`, `WaitList`, `MockFetchPlan`, short timeouts.

## Steps

1. Set `Mode = ModeWaitReattach`.
2. Default short poll interval for fast tests.

## Context

- No real daemon; fetch is fully injected via MockFetchPlan.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeWaitReattach
	if req.PollInterval <= 0 {
		req.PollInterval = 5 * time.Millisecond
	}
	return nil
}
```
