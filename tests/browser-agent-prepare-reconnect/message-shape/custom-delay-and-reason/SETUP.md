# Scenario

**Feature**: explicit delay_ms and reason appear in payload

```
BuildPrepareReconnectPayload({DelayMS: 2500, Reason: "daemon-upgrade"})
  -> delay_ms=2500; reason="daemon-upgrade"
```

## Preconditions

- MessageShapeCase = custom-delay-and-reason.
- DelayMS = 2500; Reason = `daemon-upgrade`.

## Steps

1. Set MessageShapeCase, DelayMS, Reason.

## Context

- Typical EnsureDaemon upgrade reason string (Phase 3 will pass this through).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.MessageShapeCase = MessageShapeCustomDelay
	req.DelayMS = 2500
	req.Reason = "daemon-upgrade"
	return nil
}
```
