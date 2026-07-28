# Scenario

**Feature**: optional retry hints appear when RetryBaseMS / RetryMaxMS are set

```
BuildPrepareReconnectPayload({DelayMS: 1000, RetryBaseMS: 50, RetryMaxMS: 500})
  -> retry_base_ms=50; retry_max_ms=500; delay_ms=1000
```

## Preconditions

- MessageShapeCase = with-retry-hints.
- DelayMS = 1000; RetryBaseMS = 50; RetryMaxMS = 500.

## Steps

1. Set MessageShapeCase and retry fields.

## Context

- Hints let extensions switch to aggressive reconnect backoff after prepare_reconnect.
- Keys are snake_case: `retry_base_ms`, `retry_max_ms`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.MessageShapeCase = MessageShapeWithRetryHints
	req.DelayMS = 1000
	req.RetryBaseMS = 50
	req.RetryMaxMS = 500
	req.Reason = "daemon-upgrade"
	return nil
}
```
