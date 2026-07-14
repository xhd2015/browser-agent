# Scenario

**Feature**: POST job on A — only A socket gets `type=job`

```
WS on sess-p4-a and sess-p4-b (hello on both)
POST /v1/jobs session_id=sess-p4-a
A receives job; B does not within probe window
```

## Preconditions

- Both extensions connected via per-session WS dials.

## Steps

1. Set `JobTargetSessionID = SessionIDA`.

## Context

- Mirrors browser-agent D2 but scoped to multi-session routing.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobTargetSessionID = req.SessionIDA
	return nil
}
```