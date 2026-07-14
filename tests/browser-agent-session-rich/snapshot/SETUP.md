# Scenario

**Feature**: enriched `GET /v1/session` snapshot JSON fields

```
RunDaemon -> POST /v1/sessions
GET /v1/session?session=ID -> created_at, status, session_page_count
```

## Preconditions

- Mode is `snapshot`.
- Daemon running with one created session.

## Steps

1. Set `Mode = ModeSnapshot`.
2. Leaves set `SnapshotOp`.

## Context

- Probes raw API JSON (not CLI).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSnapshot
	return nil
}
```