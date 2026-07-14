# Scenario

**Feature**: GET `/v1/session?session=<known>` → 200 + disconnected hint

```
registry pre-create sess-known
GET /v1/session?session=sess-known -> 200 waiting + hint mentions /go?session=
```

## Preconditions

- Session pre-created in registry before probe.
- No extension connected.

## Steps

1. Set `SessionID = "sess-known"`.
2. Set `PreCreateSessionIDs = []string{"sess-known"}`.

## Context

- Hint must mention keeping `/go?session=sess-known` open (phase 3 contract).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-known"
	req.PreCreateSessionIDs = []string{"sess-known"}
	return nil
}
```