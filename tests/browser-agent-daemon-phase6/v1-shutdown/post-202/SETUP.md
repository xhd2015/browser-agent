# Scenario

**Feature**: shutdown endpoint acknowledges request

```
RunDaemon -> POST /v1/shutdown -> 202 Accepted
```

## Preconditions

- `V1ShutdownOp` post-202.

## Steps

1. Set `V1ShutdownOp = V1ShutdownOpPost202`.

## Context

- Status only; daemon may still be running until harness cleanup.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.V1ShutdownOp = V1ShutdownOpPost202
	return nil
}
```