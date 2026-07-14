# Scenario

**Feature**: mock extension hellos and reaches recording before ready-timeout

```
# Happy ready path
Mock Extension -> POST /v1/hello
Control Server -> command start (via long-poll)
Mock Extension -> POST /v1/status {state: recording}
Session status: waiting_extension -> recording
```

## Preconditions

- Mock will perform hello + wait for `start` + status `recording`.
- Ready timeout is generous enough for local mock (still short for CI).

## Steps

1. Set ready timeout to a comfortable mock window (e.g. 5s).
2. Descendants choose complete-success vs complete-timeout and stop initiator.

## Context

- After ready succeeds, stop may come from extension or CLI; complete deadline applies.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReadyTimeout = 5 * time.Second
	// Default complete window; complete-fail shortens it.
	if req.CompleteTimeout < time.Second {
		req.CompleteTimeout = 5 * time.Second
	}
	return nil
}
```
