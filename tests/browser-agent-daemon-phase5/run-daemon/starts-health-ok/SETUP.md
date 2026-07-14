# Scenario

**Feature**: RunDaemon serves healthy control plane

```
RunDaemon -> GET /v1/health -> 200
```

## Preconditions

- RunDaemonOp health-ok.

## Steps

1. Set `RunDaemonOp = RunDaemonOpHealthOK`.

## Context

- Liveness only; no sessions required.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RunDaemonOp = RunDaemonOpHealthOK
	return nil
}
```