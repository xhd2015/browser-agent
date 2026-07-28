# Scenario

**Feature**: Firefox background transport is not WebSocket-only

```
# RED (Phase 1 shape)
register -> new WebSocket(/v1/ws?session=) only

# GREEN
register -> HTTP poll (/v1/ext/hello|poll|result) as primary or fallback
  WebSocket may still exist (try-once OK); must not be the only transport
```

## Preconditions

- BackgroundSourceTarget = transport-not-ws-only.

## Steps

1. Set `BackgroundSourceTarget = BgSrcTransportNotWsOnly`.

## Context

- Explicit anti-contract leaf: `isWsOnlyTransport(src)` must be false.
- Does **not** require removing WebSocket.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcTransportNotWsOnly
	return nil
}
```
