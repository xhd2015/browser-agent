# Scenario

**Feature**: clean RunDaemon shutdown removes server.json

```
RunDaemon -> server.json present
ctx cancel -> server.json absent
```

## Preconditions

- RunDaemonOp shutdown-removes-json.

## Steps

1. Set `RunDaemonOp = RunDaemonOpShutdownRemoves`.

## Context

- Harness cancels ctx inside Run before defer cleanup.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RunDaemonOp = RunDaemonOpShutdownRemoves
	return nil
}
```