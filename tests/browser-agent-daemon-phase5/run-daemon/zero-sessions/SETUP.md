# Scenario

**Feature**: fresh RunDaemon lists zero sessions

```
RunDaemon -> GET /v1/sessions -> []
```

## Preconditions

- RunDaemonOp zero-sessions.

## Steps

1. Set `RunDaemonOp = RunDaemonOpZeroSessions`.

## Context

- Empty registry before any POST /v1/sessions.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RunDaemonOp = RunDaemonOpZeroSessions
	return nil
}
```