# Scenario

**Feature**: ready-wait heartbeat lines while still no_hello

```
# Injectable short ReadyHeartbeat (product default 5s is too slow for CI)
ReadyHeartbeat=50ms ReadyTimeout≈400ms mock silent
Lifecycle Logger -> stderr: still waiting / seconds left / stage no_hello
# Expect ≥2 heartbeat-style lines (or repeated waiting + left tokens)
```

## Preconditions

- `ReadyHeartbeat` injectable via `Config.ReadyHeartbeat` (default 5s).
- `ReadyTimeout` long enough for at least two heartbeat ticks (e.g. 350–500ms
  with 50ms interval).
- Mock remains silent (`ExtNone`).

## Steps

1. Set `ReadyHeartbeat = 50ms`.
2. Set `ReadyTimeout = 400ms` (≥ 2 intervals with margin for scheduling).
3. Run until ready deadline.

## Context

- Requirement: heartbeat every ~5s in product; tests inject short interval.
- Prefer this over sleeping 5s+ in CI.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReadyHeartbeat = 50 * time.Millisecond
	req.ReadyTimeout = 400 * time.Millisecond
	req.Quiet = false
	req.Verbose = false
	req.ExtensionScript = ExtNone
	return nil
}
```
