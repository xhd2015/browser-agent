# Scenario

**Feature**: delay_ms defaults to 1000 when DelayMS is zero/negative

```
BuildPrepareReconnectPayload({DelayMS: 0})
  -> payload["delay_ms"] == 1000
```

## Preconditions

- MessageShapeCase = default-delay-ms.
- DelayMS left at 0; Reason empty; no retry hints.

## Steps

1. Set MessageShapeCase and leave DelayMS at 0.

## Context

- Canonical default for admin path when body omits delay_ms.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.MessageShapeCase = MessageShapeDefaultDelay
	req.DelayMS = 0
	req.Reason = ""
	req.RetryBaseMS = 0
	req.RetryMaxMS = 0
	return nil
}
```
