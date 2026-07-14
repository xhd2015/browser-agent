# Scenario

**Feature**: GET `/v1/session?session=<id>` — per-session snapshot

```
Test Client -> GET /v1/session?session=<id>
Registry Control Server -> snapshot JSON | 404 | 400
```

## Preconditions

- Multi-session mode: `session` query param required.

## Steps

1. Set `Mode = ModeV1Session`.
2. Leaves pre-create known session or probe unknown/missing param.

## Context

- Disconnected snapshot should include phase `waiting_extension` (or equivalent) and hint.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1Session
	return nil
}
```