# Scenario

**Feature**: enqueue then dequeue preserves identity (FIFO base)

```
JobQueue.Enqueue(job) -> job with id
JobQueue.Dequeue() -> same id/type/params (FIFO head)
```

## Preconditions

- JobOp is enqueue-dequeue family (single or two-jobs).
- Descendants choose one vs two jobs.

## Steps

1. Default JobType to `eval` when leaf does not override.
2. Children set JobOp to single or fifo-two.

## Context

- B1 single identity; B2 two-job order under child.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.JobType == "" {
		req.JobType = "eval"
	}
	if req.JobParams == nil {
		req.JobParams = map[string]any{"code": "1+1"}
	}
	return nil
}
```
