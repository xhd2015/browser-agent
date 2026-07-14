# Scenario

**Feature**: graceful kill within timeout

```
RunDaemon running
KillExistingDaemon(baseDir, 10s) -> graceful shutdown -> health down -> meta gone
```

## Preconditions

- Daemon healthy before kill.
- Default `KillTimeout` (10s).

## Steps

1. Set `KillExistingOp = KillExistingOpGraceful`.

## Context

- No injected slow grace; shutdown should complete without force kill.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.KillExistingOp = KillExistingOpGraceful
	return nil
}
```