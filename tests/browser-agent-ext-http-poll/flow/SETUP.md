# Scenario

**Feature**: end-to-end HTTP poll transport happy path

```
hello -> connected
enqueue job (bg)
poll -> job
result -> /v1/jobs ok
```

## Preconditions

- Mode = flow.

## Steps

1. Set Mode to flow.
2. Leaves set FlowCase.

## Context

- Documents the full Phase 1 contract in one leaf.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeFlow
	if req.PollWaitMS <= 0 {
		req.PollWaitMS = 3000
	}
	return nil
}
```
