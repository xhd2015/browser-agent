# Scenario

**Feature**: force kill after graceful timeout

```
RunDaemon with slow ShutdownGracePeriod
KillExistingDaemon(baseDir, short timeout) -> SIGKILL path -> meta gone
```

## Preconditions

- Injected slow drain via `ShutdownGracePeriod`.
- Short `KillTimeout` shorter than grace period.

## Steps

1. Set `KillExistingOp = KillExistingOpForce`.
2. Set `ShutdownGracePeriod = 3s`.
3. Set `KillTimeout = 150ms`.

## Context

- Grace period exceeds kill wait → implementer must force-kill pid (or equivalent
  in-process cancel) and remove `server.json`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.KillExistingOp = KillExistingOpForce
	req.ShutdownGracePeriod = 3 * time.Second
	req.KillTimeout = 150 * time.Millisecond
	return nil
}
```