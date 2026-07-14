# Scenario

**Feature**: Wait times out when never completed (B4)

```
Enqueue(job) with short timeout
Wait(ctx short) without Complete
  -> ok=false, error contains "timeout"
```

## Preconditions

- No Complete call.
- JobTimeout ~80ms.

## Steps

1. Set `JobOp = JobOpWaitTimeout`.
2. Set short `JobTimeout`.

## Context

- Distinct from expire leaf: here primary assert is Wait error/result text.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpWaitTimeout
	req.JobTimeout = 80 * time.Millisecond
	return nil
}
```
