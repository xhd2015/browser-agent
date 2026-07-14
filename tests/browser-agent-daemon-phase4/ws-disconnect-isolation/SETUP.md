# Scenario

**Feature**: WS disconnect fails inflight jobs on the disconnected session only

```
Fake Extension hello on A and B
POST /v1/jobs on A (inflight)
close A WS -> A job fails (disconnect policy v1)
B remains connected and can complete jobs
```

## Preconditions

- Two sessions with independent WS sockets.
- Mode is `ws-disconnect-isolation`.

## Steps

1. Set `Mode = ModeWSDisconnectIsolation`.

## Context

- Disconnect policy v1: fail inflight, no requeue (same as browser-agent D4).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWSDisconnectIsolation
	return nil
}
```