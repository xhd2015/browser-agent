# Scenario

**Feature**: no extension + short timeout → result error timeout (C2)

```
No Fake Extension connected
Test Client -> POST /v1/jobs timeout_ms=150
  -> HTTP 200 (or 408/504) with ok=false and error containing timeout
```

## Preconditions

- FakeExtension false.
- Short JobHTTPTimeoutMS.

## Steps

1. Disable fake extension.
2. Set timeout_ms to 150.

## Context

- Product may return 200 with JobResult ok=false (preferred enqueue-and-wait style)
  or a gateway timeout status; assert focuses on timeout signal + not-ok.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FakeExtension = false
	req.FakeExtensionResult = false
	req.JobHTTPTimeoutMS = 150
	return nil
}
```
