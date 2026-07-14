# Scenario

**Feature**: single enqueue/dequeue round-trip (B1)

```
Enqueue(type=eval, params={code:1+1}) -> Dequeue
  same id, type eval, params preserved
```

## Preconditions

- Exactly one job.

## Steps

1. Set `JobOp = JobOpEnqueueDequeue` (`"enqueue-dequeue"`).

## Context

- After dequeue, status is typically `running` (worker claimed it).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpEnqueueDequeue
	req.JobType = "eval"
	req.JobParams = map[string]any{"code": "1+1"}
	return nil
}
```
