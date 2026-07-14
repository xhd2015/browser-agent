# Scenario

**Feature**: Job push delivered only on the target session's WebSocket

```
Fake Extension hello on A and B (both sockets open)
POST /v1/jobs session_id=A
only A socket receives type=job; B socket must not
```

## Preconditions

- Two sessions with independent WS sockets.
- Mode is `ws-job-routing`.

## Steps

1. Set `Mode = ModeWSJobRouting`.
2. Leaf sets job target session A.

## Context

- Cross-session job leak is a hard failure.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWSJobRouting
	return nil
}
```