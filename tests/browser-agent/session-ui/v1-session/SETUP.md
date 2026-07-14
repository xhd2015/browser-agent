# Scenario

**Feature**: GET /v1/session JSON for browser-agent page poller

```
Test Client -> GET /v1/session?session=<id>
Control Server -> {session_id, phase, extension.connected, supports_browser_agent, hint}
```

## Preconditions

- Probe is v1-session.
- Known live session id (root Setup).

## Steps

1. Set `Probe = ProbeV1Session`.
2. Children choose DoWSHello.

## Context

- E1 no hello vs E2 after hello.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbeV1Session
	return nil
}
```
