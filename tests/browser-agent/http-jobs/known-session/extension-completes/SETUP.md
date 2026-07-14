# Scenario

**Feature**: fake WS extension completes job → HTTP 200 ok (C1)

```
Fake Extension dials /v1/ws, hello, auto result ok
Test Client -> POST /v1/jobs type=eval wait
  -> HTTP 200, JobResult ok=true
```

## Preconditions

- FakeExtension true with FakeExtensionResult true.
- Generous timeout_ms (3s).

## Steps

1. Enable fake extension auto-complete.
2. Set JobHTTPTimeoutMS to 3000.

## Context

- No real CDP; extension is harness goroutine.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FakeExtension = true
	req.FakeExtensionResult = true
	req.JobHTTPTimeoutMS = 3000
	req.JobHTTPType = "eval"
	req.JobHTTPParams = map[string]any{"code": "1+1"}
	return nil
}
```
