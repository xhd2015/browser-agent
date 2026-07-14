# Scenario

**Feature**: plain serve does not invoke chrome hooks

```
HandleCLI(serve --host --port --base-dir) -> health ok; LaunchFn 0; OpenChromeFn 0
```

## Preconditions

- Ephemeral port; short-lived daemon in Run.

## Steps

1. Set `ServeNoChromeOp = run-daemon-no-launch`.

## Context

- RunDaemon path only (no --session-id legacy Run).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeNoChromeOp = ServeNoChromeOpRunDaemonNoLaunch
	return nil
}
```