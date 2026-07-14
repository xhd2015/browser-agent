# Scenario

**Feature**: extension type=result unblocks HTTP waiter (D3)

```
Fake Extension hello + AutoCompleteOK Loop
Test Client -> POST /v1/jobs wait
  <- JobResult ok=true when WS result arrives
```

## Preconditions

- WSAction result-unblocks.
- Extension completes first job with ok.

## Steps

1. Set WSAction to result-unblocks.

## Context

- Complements C1 but asserts the WS result path explicitly under ws-control.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSAction = WSActionResultUnblocks
	return nil
}
```
