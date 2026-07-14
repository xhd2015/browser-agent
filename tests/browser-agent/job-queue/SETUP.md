# Scenario

**Feature**: in-memory JobQueue FIFO, wait/complete, timeout/expire

```
# Pure package API — no Control Server, no Chrome
Test Client -> NewJobQueue()
  -> Enqueue(Job) -> Dequeue / Wait / Complete / Get
Job status: queued | running | done | failed | expired
```

## Preconditions

- Mode is pure job-queue (`ModeJobQueue`).
- Package exports `NewJobQueue` and job types used by harness Run.
- No listen socket.

## Steps

1. Set `Mode = ModeJobQueue`.
2. Descendants set `JobOp` and job fields.

## Context

- Requirement B1–B6. Integration with HTTP/WS is under `http-jobs/` and `ws-control/`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeJobQueue
	return nil
}
```
