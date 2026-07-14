# Scenario

**Feature**: GET `/v1/sessions` — list session snapshots

```
Test Client -> GET /v1/sessions -> JSON array sorted by session_id
```

## Preconditions

- Mode `ModeV1SessionsGet`.

## Steps

1. Set `Mode = ModeV1SessionsGet`.
2. Leaves choose empty vs two sessions.

## Context

- Pre-create uses POST `/v1/sessions` in Run (wire-level, not direct registry API).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1SessionsGet
	return nil
}
```