# Scenario

**Feature**: Complete on unknown job id returns error (B5)

```
NewJobQueue (empty)
Complete("no-such-job-id", ...) -> error
```

## Preconditions

- Queue has never enqueued that id.

## Steps

1. Set `JobOp = JobOpCompleteUnknown`.

## Context

- Documented alternative was no-op; MVP picks **error** for unknown id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpCompleteUnknown
	return nil
}
```
