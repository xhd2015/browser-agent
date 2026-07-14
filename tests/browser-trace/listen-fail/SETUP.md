# Scenario

**Feature**: Control Server fails to bind — hard error, no fallback port

```
# Another process already owns Addr
Occupier -> Listen(Addr)  # test holds the port

# browser-trace attempts the same address and must fail immediately
User -> browser-trace.Run(Addr) -> bind error (exit ≠ 0)
# no alternate port, no silent success
```

## Preconditions

- `req.OccupyAddr` will be true so a real TCP listener holds `Addr` before start.
- Extension script remains `none` (server never becomes ready).

## Steps

1. Mark this branch as listen-failure coverage.
2. Descendants set `OccupyAddr` and a concrete `Addr`.

## Context

- Product policy: fixed default `127.0.0.1:43759`; tests prove “busy ⇒ hard fail”
  on an ephemeral address with the same bind logic.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtNone
	req.StopMode = StopNone
	req.OccupyAddr = true
	// Short timeouts irrelevant if bind fails first, but keep small for safety.
	req.ReadyTimeout = 500 * time.Millisecond
	req.CompleteTimeout = 500 * time.Millisecond
	return nil
}
```
