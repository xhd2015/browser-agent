# Scenario

**Feature**: `GET /v1/session` includes `created_at`

```
POST /v1/sessions -> GET /v1/session?session=ID
```

## Preconditions

- `SnapshotOp = created-at-in-json`.
- No extension connection required.

## Steps

1. Set `SnapshotOp = SnapshotOpCreatedAt`.
2. Set `SessionID = sess-rich-created`.

## Context

- `created_at` must be non-empty RFC3339-ish timestamp in JSON.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SnapshotOp = SnapshotOpCreatedAt
	req.SessionID = "sess-rich-created"
	return nil
}
```