# Scenario

**Feature**: Complete unblocks Wait with equal payload (B3)

```
Enqueue -> goroutine Wait(jobID)
Complete(jobID, {ok:true, data:{value:2}})
Wait returns same ok/data; job status done
```

## Preconditions

- JobOp complete-unblocks.
- CompleteOK true with data value 2.

## Steps

1. Set `JobOp = JobOpCompleteUnblocks`.
2. Set CompleteOK and CompleteData.

## Context

- Harness starts Wait before Complete with a short yield.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpCompleteUnblocks
	req.CompleteOK = true
	req.CompleteData = map[string]any{"value": float64(2)}
	return nil
}
```
