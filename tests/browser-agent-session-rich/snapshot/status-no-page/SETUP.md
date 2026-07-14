# Scenario

**Feature**: page count 0 → `status=no_session_page`

```
Fake Extension -> hello { session_page_count: 0 }
GET /v1/session -> status no_session_page
```

## Preconditions

- `SnapshotOp = status-no-page`.
- Hello reports zero session pages.

## Steps

1. Set `SnapshotOp = SnapshotOpStatusNoPage`.
2. Set `SessionID = sess-rich-no-page`.

## Context

- count==0 → `no_session_page` regardless of connection.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SnapshotOp = SnapshotOpStatusNoPage
	req.SessionID = "sess-rich-no-page"
	return nil
}
```