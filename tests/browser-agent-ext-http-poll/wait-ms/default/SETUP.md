# Scenario

**Feature**: wait_ms <= 0 normalizes to default 25000

```
NormalizeExtPollWaitMS(0) -> 25000
```

## Preconditions

- WaitMSCase = default.
- WaitMSInput = 0.

## Steps

1. Set case and input 0.

## Context

- Also covers "unset" semantic for omitted wait_ms handlers using 0.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.WaitMSCase = WaitMSDefault
	req.WaitMSInput = 0
	return nil
}
```

