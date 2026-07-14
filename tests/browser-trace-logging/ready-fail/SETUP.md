# Scenario

**Feature**: ready-phase failure — logging around wait and timeout

```
# Ready never becomes recording within ReadyTimeout
Control Server waiting_extension
(time passes ReadyTimeout)
Lifecycle Logger -> stderr: heartbeats (if wait long enough) + rich timeout
browser-trace -> exit ≠ 0
```

## Preconditions

- Mock does not reach recording (silent or incomplete).
- Short ready timeouts unless a leaf needs multiple heartbeats.

## Steps

1. Default extension script to none; leaves may keep that.
2. Shorten CompleteTimeout (unused on ready-fail but harmless).

## Context

- Focus: ready-fail stderr richness and heartbeats, not HAR save.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtNone
	req.StopMode = StopNone
	req.Verbose = false
	req.Quiet = false
	// Leaves set ReadyTimeout / ReadyHeartbeat as needed.
	if req.CompleteTimeout == 0 || req.CompleteTimeout >= 30*time.Second {
		req.CompleteTimeout = 1 * time.Second
	}
	return nil
}
```
