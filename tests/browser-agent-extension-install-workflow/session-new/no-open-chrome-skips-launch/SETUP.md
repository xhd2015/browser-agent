# Scenario

**Feature**: --no-open-chrome skips launch but still extracts canonical path

```
SessionNew(NoOpenChrome=true) -> LaunchFn 0; canonical path still on disk
```

## Preconditions

- `NoOpenChrome = true`.

## Steps

1. Set `SessionNewOp = no-open-chrome-skips-launch`.
2. Set `NoOpenChrome = true`.

## Context

- Session still created on daemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpNoOpenChromeSkipsLaunch
	req.NoOpenChrome = true
	return nil
}
```