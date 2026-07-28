# Scenario

**Feature**: HTTP-connected session receives prepare_reconnect via poll events

```
Create + hello sess-ext-poll-prep
BroadcastPrepareReconnect(registry, {reason:"daemon-upgrade", delay_ms:1000})
POST /v1/ext/poll {session_id, wait_ms:500}
  -> 200 events include type prepare_reconnect
```

## Preconditions

- PollCase = prepare-reconnect-event.
- BroadcastPrepareReconnect = true.
- No job enqueue.

## Steps

1. Hello (HTTP attach, no WS).
2. Broadcast prepare_reconnect on registry.
3. Poll; expect event delivery (not WS-only).

## Context

- Implementer must queue poll events for HTTP-connected sessions; WS write alone is insufficient.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PollCase = PollPrepareReconnectEvent
	req.SessionID = "sess-ext-poll-prep"
	req.BroadcastPrepareReconnect = true
	req.BroadcastReason = "daemon-upgrade"
	req.BroadcastDelayMS = 1000
	req.PollWaitMS = 500
	req.EnqueueJob = false
	return nil
}
```
