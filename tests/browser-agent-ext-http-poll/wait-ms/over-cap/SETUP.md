# Scenario

**Feature**: wait_ms above max is clamped to 30000

```
NormalizeExtPollWaitMS(99999) -> 30000
```

## Preconditions

- WaitMSCase = over-cap.
- WaitMSInput = 99999.

## Steps

1. Set case and oversized input.

## Context

- Cap protects daemon from multi-minute long-poll holds.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitMSCase = WaitMSOverCap
	req.WaitMSInput = 99999
	return nil
}
```
