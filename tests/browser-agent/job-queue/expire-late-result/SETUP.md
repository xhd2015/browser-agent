# Scenario

**Feature**: waiter timeout expires job; late Complete is safe (B6)

```
Enqueue + Wait(short) without worker Complete
  -> job status expired (or failed-timeout)
Late Complete(ok) must not panic; status must not become successful done via late ok
```

## Preconditions

- Short JobTimeout.
- Late Complete attempted after Wait returns.

## Steps

1. Set `JobOp = JobOpExpireLate`.
2. Set JobTimeout 80ms.

## Context

- Late result ignored **or** returns error — both OK; must not flip to done/ok success.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobOp = JobOpExpireLate
	req.JobTimeout = 80 * time.Millisecond
	return nil
}
```
