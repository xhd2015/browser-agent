# Scenario

**Feature**: wait_ms within cap is unchanged

```
NormalizeExtPollWaitMS(5000) -> 5000
```

## Preconditions

- WaitMSCase = within-cap.
- WaitMSInput = 5000 (under MaxExtPollWaitMS).

## Steps

1. Set case and input 5000.

## Context

- Tests use short wait_ms (e.g. 100–5000) so leaves finish quickly.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitMSCase = WaitMSWithinCap
	req.WaitMSInput = 5000
	return nil
}
```
