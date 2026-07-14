# Scenario

**Feature**: two jobs dequeue in FIFO order (B2)

```
Enqueue(A=eval) then Enqueue(B=info)
Dequeue -> A; Dequeue -> B
```

## Preconditions

- First job type eval; second info.

## Steps

1. Set `JobOp = JobOpFIFOTwo`.
2. Set first type `eval`, second `info`.

## Context

- Order is the primary assertion (ids may be any unique strings).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpFIFOTwo
	req.JobType = "eval"
	req.JobParams = map[string]any{"code": "A"}
	req.SecondJobType = "info"
	req.SecondParams = map[string]any{}
	return nil
}
```
